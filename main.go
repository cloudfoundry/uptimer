package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/uptimer"
)

func main() {
	configPath := flag.String("configFile", "", "Path to the config file")
	flag.Parse()
	if *configPath == "" {
		panic("configPath required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	runner := cmdRunner.New(os.Stdout, os.Stderr, io.Copy)
	cfCmdGenerator := cfCmdGenerator.New()
	workflow := cfWorkflow.New(
		cfg.Api,
		cfg.AdminUsername,
		cfg.AdminUsername,
		"org",
		"space",
		"appName",
		"appPath",
		cfg.SkipSslValidation, cfCmdGenerator,
	)

	uptimer := uptimer.New(cfg, logger, workflow, runner)

	if err := uptimer.Setup(); err != nil {
		logger.Fatalln("Failed setup:", err)
	}

	defer func() {
		if err := uptimer.TearDown(); err != nil {
			logger.Fatalln("Failed teardown:", err)
		}
	}()

	if err := uptimer.Run(); err != nil {
		logger.Fatalln("Failed run:", err)
	}
}
