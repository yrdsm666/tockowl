package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/utils"
)

var logging *logrus.Logger

func init() {
	fcfg, err := config.NewConfig("config/config.yaml", 0, "")
	var cfg *config.LogConfig
	if err == nil {
		cfg = fcfg.Log
	} else {
		// for testing
		cfg = &config.LogConfig{
			Level:  "debug",
			ToFile: false,
		}
	}
	level, err := logrus.ParseLevel(cfg.Level)
	utils.PanicOnError(err)
	var out io.Writer
	if cfg.ToFile {
		file, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY, 0644)
		utils.PanicOnError(err)
		out = io.MultiWriter(os.Stdout, file)
	} else {
		out = os.Stdout
	}
	logging = &logrus.Logger{
		Out: out,
		Formatter: &logrus.TextFormatter{
			ForceColors:     true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
		},
		Level: level,
	}
}

// should be called after InitLog
func GetLogger() *logrus.Logger {
	return logging
}
