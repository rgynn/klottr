package main

import (
	"github.com/rgynn/klottr/cmd/intg_test/internal/tester"
	"github.com/rgynn/klottr/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {

	logger := logrus.New()

	cfg, err := config.NewFromEnv()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	tstr, err := tester.NewTester(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}

	if err := tstr.Run(); err != nil {
		logger.Fatal(err)
	}
}
