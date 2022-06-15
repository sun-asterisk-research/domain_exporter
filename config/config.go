package config

import (
	"io/ioutil"
	"os"

	"github.com/cloudprober/cloudprober/config"
	"github.com/sirupsen/logrus"
)

func readFile(fileName string) string {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		logrus.Fatalf("Failed to read the config file: %v", err)
	}

	return string(b)
}

func GetConfig(path string) string {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return readFile(path)
	}

	logrus.Warningf("Config file %s not found. Using default config.", path)

	return config.DefaultConfig()
}
