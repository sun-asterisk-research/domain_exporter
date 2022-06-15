package logging

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type config struct {
	level  string
	path   string
	format string
}

func getConfig(flags *pflag.FlagSet) config {
	var conf config

	debug, _ := flags.GetBool("debug")
	if debug {
		conf.level = "debug"
	} else {
		conf.level, _ = flags.GetString("log-level")
	}

	conf.path, _ = flags.GetString("log-path")
	conf.format, _ = flags.GetString("log-format")

	return conf
}

// Configure set logrus options
func Configure(flags *pflag.FlagSet) {
	config := getConfig(flags)

	loglevel, err := logrus.ParseLevel(config.level)
	if err == nil {
		logrus.SetLevel(loglevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if config.path != "" {
		if file, err := os.OpenFile(config.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
			logrus.Error(err)
		} else {
			logrus.SetOutput(file)
		}
	}

	if loglevel == logrus.DebugLevel {
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		switch config.format {
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		default:
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
			})
		}
	}
}
