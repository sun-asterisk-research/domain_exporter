package main

import (
	"github.com/sun-asterisk-research/domain_exporter/logging"
	"github.com/sun-asterisk-research/domain_exporter/prober"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "domain_exporter",
	Short: "Performs WHOIS lookups for a list of probe provided in the config file and exposes them",
	Run:   run,
}

func main() {
	flags := cmd.Flags()

	flags.StringP("config-file", "C", "/etc/domain_exporter.cfg", "Set the config file")
	flags.BoolP("debug", "D", false, "Enable debug mode")
	flags.StringP("log-level", "l", "info", "Set the logging level")
	flags.String("log-format", "text", "Set the logging format")
	flags.String("log-path", "", "Set the logging output file")

	cmd.Execute()
}

func run(cmd *cobra.Command, args []string) {
	logging.Configure(cmd.Flags())

	configFile, _ := cmd.Flags().GetString("config-file")

	prober.Start(configFile)

	select {}
}
