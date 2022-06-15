package prober

import (
	"github.com/cloudprober/cloudprober/probes"
	"github.com/sun-asterisk-research/domain_exporter/probes/domain"
	domainpb "github.com/sun-asterisk-research/domain_exporter/probes/domain/proto"
)

func registerProbes() {
	probes.RegisterProbeType(int(domainpb.E_DomainProbe.TypeDescriptor().Number()), func() probes.Probe { return &domain.Probe{} })
}
