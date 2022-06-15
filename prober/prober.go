package prober

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudprober/cloudprober"
	"github.com/cloudprober/cloudprober/web"
	"github.com/sirupsen/logrus"
	"github.com/sun-asterisk-research/domain_exporter/config"
)

func init() {
	registerProbes()
}

type Prober struct {
	configFile string
	cancel     context.CancelFunc
}

func (p *Prober) Start() {
	var serveMux http.ServeMux

	http.DefaultServeMux = &serveMux

	configStr := config.GetConfig(p.configFile)
	err := cloudprober.InitFromConfig(configStr)
	if err != nil {
		logrus.Fatalf("Could not load config: %v", err)
	}

	web.Init()

	http.HandleFunc("/-/reload", p.handleReload)

	ctx, cancel := context.WithCancel(context.Background())
	cloudprober.Start(ctx)

	p.cancel = cancel
}

func (p *Prober) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
		return
	}

	logrus.Info("Config reload requested")

	p.cancel()
	// Wait for cloudprober to fully shut down
	// Cloudprober functions share a lock so may be this is enough
	// There's also a test case for this so it's probably safe to do this
	time.Sleep(time.Second)
	p.Start()
}

var prober = &Prober{}

func Start(configFile string) {
	prober.configFile = configFile
	prober.Start()
}
