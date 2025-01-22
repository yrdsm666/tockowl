package main

import (
	"os"

	"github.com/yrdsm666/tockowl/cmd/clientcore"
	"github.com/yrdsm666/tockowl/logging"
)

var logger = logging.GetLogger()

func main() {
	logger.SetOutput(os.Stdout) // only output to stdout
	client := clientcore.NewClient(logger.WithField("module", "client"))
	if len(os.Args) == 1 {
		client.NormalRun()
	} else {
		logger.Infof("wrong number of args")
	}
	client.Stop()
}
