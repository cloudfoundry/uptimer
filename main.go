package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/measurement"
	"github.com/cloudfoundry/uptimer/orchestrator"
)

func main() {
	configPath := flag.String("configFile", "", "Path to the config file")
	flag.Parse()
	if *configPath == "" {
		log.Fatalln("Error: '-configFile' flag required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC)
	runner := cmdRunner.New(os.Stdout, os.Stderr, io.Copy)
	cfCmdGenerator := cfCmdGenerator.New()
	workflow := cfWorkflow.New(cfg.CF, cfCmdGenerator)
	availabilityMeasurement := measurement.NewAvailability(
		workflow.AppUrl(),
		time.Second,
		clock.New(),
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	)

	orc := orchestrator.New(cfg.While, logger, workflow, runner, []measurement.Measurement{availabilityMeasurement})

	logger.Println("Setting up")
	if err := orc.Setup(); err != nil {
		logger.Println("Failed setup:", err)
		TearDownAndExit(orc, logger)
	}

	if err := orc.Run(); err != nil {
		logger.Println("Failed run:", err)
		TearDownAndExit(orc, logger)
	}

	TearDownAndExit(orc, logger)
}

func TearDownAndExit(orc orchestrator.Orchestrator, logger *log.Logger) {
	logger.Println("Tearing down")
	if err := orc.TearDown(); err != nil {
		logger.Fatalln("Failed teardown:", err)
	}
}
