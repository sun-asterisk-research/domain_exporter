package domain

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cloudprober/cloudprober/metrics"
	"github.com/cloudprober/cloudprober/probes/options"
	"github.com/cloudprober/cloudprober/targets/endpoint"
	"github.com/domainr/whois"
	"github.com/sirupsen/logrus"
	"github.com/sun-asterisk-research/domain_exporter/common"
	"github.com/sun-asterisk-research/domain_exporter/probes"
	configpb "github.com/sun-asterisk-research/domain_exporter/probes/domain/proto"
)

// DefaultTargetsUpdateInterval defines default frequency for target updates.
// Actual targets update interval is:
// max(DefaultTargetsUpdateInterval, probe_interval)
var (
	DefaultTargetsUpdateInterval = 12 * time.Hour
	expiryRegex                  = regexp.MustCompile(`(?i)(\[有効期限]|Registry Expiry Date|paid-till|Expiration Date|Expiration Time|Expiry.*|expires.*|expire-date)[?:|\s][ \t](.*)`)

	formats = []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"02-Jan-2006",
		"2006.01.02",
		"Mon Jan 2 15:04:05 MST 2006",
		"02/01/2006",
		"2006-01-02 15:04:05 MST",
		"2006/01/02",
		"Mon Jan 2006 15:04:05",
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05",
		"2.1.2006 15:04:05",
		"02/01/2006 15:04:05",
	}
)

// Probe holds aggregate information about all probe runs, per-target.
type Probe struct {
	name   string
	opts   *options.Options
	config *configpb.ProbeConf
	logger *logrus.Entry

	// book-keeping params
	targets []endpoint.Endpoint
	scheme  string

	// How often to resolve targets (in probe counts), it's the minimum of
	targetsUpdateInterval time.Duration

	// Cancel functions for per-target probe loop
	cancelFuncs map[string]context.CancelFunc
	waitGroup   sync.WaitGroup

	requestBody []byte
}

// Init initializes the probe with the given params.
func (p *Probe) Init(name string, opts *options.Options) error {
	httpConfig, ok := opts.ProbeConf.(*configpb.ProbeConf)
	if !ok {
		return fmt.Errorf("not domain config")
	}

	p.name = name
	p.opts = opts
	p.config = httpConfig

	p.logger = logrus.WithFields(logrus.Fields{
		"name":  p.name,
		"probe": "domain",
	})

	p.targets = p.opts.Targets.ListEndpoints()
	p.cancelFuncs = make(map[string]context.CancelFunc, len(p.targets))

	p.targetsUpdateInterval = DefaultTargetsUpdateInterval
	// There is no point refreshing targets before probe interval.
	if p.targetsUpdateInterval < p.opts.Interval {
		p.targetsUpdateInterval = p.opts.Interval
	}

	p.logger.Infof("Targets update interval: %v", p.targetsUpdateInterval)

	return nil
}

func (p *Probe) GetName() string {
	return p.name
}

func (p *Probe) GetOpts() *options.Options {
	return p.opts
}

func (p *Probe) GetType() string {
	return "domain"
}

func (p *Probe) Logger() *logrus.Entry {
	return p.logger
}

func (p *Probe) Run(ctx context.Context, target endpoint.Endpoint, em *metrics.EventMetrics) (success bool, err error) {
	logger := p.logger.WithField("target", target.Name)

	logger.Infof("Domain: %v", p.config.Domain)

	em.AddLabel("domain", *p.config.Domain)

	dateExpiry := metrics.NewInt(0)
	em.AddMetric("probe_domain_expiration", dateExpiry)

	ctx, cancel := context.WithTimeout(ctx, p.opts.Timeout)

	defer cancel()

	date, err := lookup(*p.config.Domain)

	if err != nil {
		logrus.WithField("domain", &p.config.Domain).Errorf("Probe failed with err: %v", err)
		return false, nil
	}

	dateExpiry.IncBy(metrics.NewInt(date))

	return true, nil
}

func lookup(domain string) (int64, error) {
	req, err := whois.NewRequest(domain)
	if err != nil {
		return -1, err
	}

	res, err := whois.DefaultClient.Fetch(req)
	if err != nil {
		return -1, err
	}

	date, err := parse(domain, res.Body)

	return date, nil
}

func parse(host string, res []byte) (int64, error) {
	results := expiryRegex.FindStringSubmatch(string(res))
	if len(results) < 1 {
		err := fmt.Errorf("don't know how to parse domain: %s", host)
		return -2, err
	}

	for _, format := range formats {
		if date, err := time.Parse(format, strings.TrimSpace(results[2])); err == nil {
			logrus.WithFields(logrus.Fields{
				"host":      host,
				"date":      date,
				"date unix": date.Unix(),
			}).Info("Domain expiration")
			return date.Unix(), nil
		}

	}
	return -1, errors.New(fmt.Sprintf("Unable to parse date: %s, for %s\n", strings.TrimSpace(results[2]), host))
}

// updateTargetsAndStartProbes refreshes targets and starts probe loop for
// new targets and cancels probe loops for targets that are no longer active.
// Note that this function is not concurrency safe. It is never called
// concurrently by Start().
func (p *Probe) updateTargetsAndStartProbes(ctx context.Context, dataChan chan *metrics.EventMetrics) {
	p.targets = p.opts.Targets.ListEndpoints()

	p.logger.Debugf("Probe(%s) got %d targets", p.name, len(p.targets))

	// updatedTargets is used only for logging.
	updatedTargets := make(map[string]string)
	defer func() {
		if len(updatedTargets) > 0 {
			p.logger.Infof("Probe(%s) targets updated: %v", p.name, updatedTargets)
		}
	}()

	activeTargets := make(map[string]endpoint.Endpoint)
	for _, target := range p.targets {
		key := target.Key()
		activeTargets[key] = target
	}

	// Stop probing for deleted targets by invoking cancelFunc.
	for targetKey, cancelF := range p.cancelFuncs {
		if _, ok := activeTargets[targetKey]; ok {
			continue
		}
		cancelF()
		updatedTargets[targetKey] = "DELETE"
		delete(p.cancelFuncs, targetKey)
	}

	gapBetweenTargets := 10 * time.Millisecond
	var startWaitTime time.Duration

	// Start probe loop for new targets.
	for key, target := range activeTargets {
		// This target is already initialized.
		if _, ok := p.cancelFuncs[key]; ok {
			continue
		}
		updatedTargets[key] = "ADD"

		probeCtx, cancelF := context.WithCancel(ctx)
		p.waitGroup.Add(1)

		go func(target endpoint.Endpoint, waitTime time.Duration) {
			defer p.waitGroup.Done()
			// Wait for wait time + some jitter before starting this probe loop.
			time.Sleep(waitTime + time.Duration(rand.Int63n(gapBetweenTargets.Microseconds()/10))*time.Microsecond)
			probes.RunProbe(probeCtx, p, target, dataChan)
		}(target, startWaitTime)

		startWaitTime += gapBetweenTargets

		p.cancelFuncs[key] = cancelF
	}
}

// wait waits for child go-routines (one per target) to clean up.
func (p *Probe) wait() {
	p.waitGroup.Wait()
}

// Start starts and runs the probe indefinitely.
func (p *Probe) Start(ctx context.Context, dataChan chan *metrics.EventMetrics) {
	defer p.wait()

	p.updateTargetsAndStartProbes(ctx, dataChan)

	// Do more frequent listing of targets until we get a non-zero list of
	// targets.
	initialRefreshInterval := p.opts.Interval
	// Don't wait too long if p.opts.Interval is large.
	if initialRefreshInterval > time.Second {
		initialRefreshInterval = time.Second
	}

	for {
		if common.IsCtxDone(ctx) {
			return
		}
		if len(p.targets) != 0 {
			break
		}
		p.updateTargetsAndStartProbes(ctx, dataChan)
		time.Sleep(initialRefreshInterval)
	}

	targetsUpdateTicker := time.NewTicker(p.targetsUpdateInterval)
	defer targetsUpdateTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-targetsUpdateTicker.C:
			p.updateTargetsAndStartProbes(ctx, dataChan)
		}
	}
}
