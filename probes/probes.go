package probes

import (
	"context"
	"time"

	"github.com/cloudprober/cloudprober/metrics"
	"github.com/cloudprober/cloudprober/probes/options"
	"github.com/cloudprober/cloudprober/targets/endpoint"
	"github.com/sirupsen/logrus"
	"github.com/sun-asterisk-research/domain_exporter/common"
)

type Probe interface {
	GetName() string
	GetOpts() *options.Options
	GetType() string
	Logger() *logrus.Entry
	Run(ctx context.Context, target endpoint.Endpoint, em *metrics.EventMetrics) (success bool, err error)
}

func RunProbe(ctx context.Context, p Probe, target endpoint.Endpoint, dataChan chan *metrics.EventMetrics) {
	opts := p.GetOpts()
	logger := p.Logger()

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for ts := time.Now(); true; ts = <-ticker.C {
		// Don't run another probe if context is canceled already.
		if common.IsCtxDone(ctx) {
			return
		}

		em := metrics.NewEventMetrics(ts)

		em.Kind = metrics.GAUGE
		em.LatencyUnit = opts.LatencyUnit
		em.AddLabel("probe", p.GetType())
		em.AddLabel("target", p.GetName())

		for _, label := range opts.AdditionalLabels {
			em.AddLabel(label.KeyValueForTarget(target.Name))
		}

		logger.Debug("Starting probe")

		start := time.Now()

		success, err := p.Run(ctx, target, em)
		if err == nil {
			if success {
				em.AddMetric("probe_domain_success", metrics.NewInt(1))
				logger.Info("Probe succeeded")
			} else {
				em.AddMetric("probe_domain_success", metrics.NewInt(0))
				logger.Info("Probe failed")
			}
		} else {
			em.AddMetric("probe_domain_success", metrics.NewInt(0))
			logger.Infof("Probe failed: %v", err)
		}

		em.AddMetric("probe_domain_duration_seconds", metrics.NewFloat(time.Since(start).Seconds()))

		opts.LogMetrics(em)

		dataChan <- em
	}
}
