# Domain Exporter

Very simple service which performs WHOIS lookups for a list of probe provided in the config file 
and exposes them on a "/metrics" endpoint for consumption via Prometheus.

```text
probe {
	name: "Google homepage"
	type: EXTENSION
	targets { dummy_targets {} }
	[domain_probe] {
	    domain: "google.com"
	}
}
```

Flags:

```text
usage: domain_exporter [<flags>]

Flags:
      --config-file=./examples/cloudprober.cfg  Set the config file
  -l  --log-level=info                          Set the logging level
      --log-format=logfmt                       Set the logging format
      --log-path=logfmt                         Set the logging output file
  -D, --debug=logfmt                            Enable debug mode
```