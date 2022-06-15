package glog

import (
	"flag"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Level int32

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	*l = Level(v)
	return nil
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

type Verbose bool

var verbosity Level

// init replicates glog's verbosity level functionality,
// allowing us to show and hide high-verbosity-level
// messages from various kubernetes components.
func init() {
	flag.Var(&verbosity, "v", "log level for V logs")
}

func V(level Level) Verbose {
	return verbosity >= level
}

func (v Verbose) Info(args ...interface{}) {
	if v {
		logrus.Debug(args...)
	}
}

func (v Verbose) Infoln(args ...interface{}) {
	if v {
		logrus.Debugln(args...)
	}
}

func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		logrus.Debugf(format, args...)
	}
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func InfoDepth(depth int, args ...interface{}) {
	logrus.Info(args...)
}

func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warning(args ...interface{}) {
	logrus.Warn(args...)
}

func WarningDepth(depth int, args ...interface{}) {
	logrus.Warn(args...)
}

func Warningln(args ...interface{}) {
	logrus.Warnln(args...)
}

func Warningf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func ErrorDepth(depth int, args ...interface{}) {
	logrus.Error(args...)
}

func Errorln(args ...interface{}) {
	logrus.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func FatalDepth(depth int, args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalln(args ...interface{}) {
	logrus.Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Exit(args ...interface{}) {
	logrus.Fatal(args...)
}

func ExitDepth(depth int, args ...interface{}) {
	logrus.Fatal(args...)
}

func Exitln(args ...interface{}) {
	logrus.Fatalln(args...)
}

func Exitf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}
