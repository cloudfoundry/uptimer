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

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	runner := cmdRunner.New(os.Stdout, os.Stderr, io.Copy)
	cfCmdGenerator := cfCmdGenerator.New()
	workflow := cfWorkflow.New(cfg.CF, cfCmdGenerator)

	orc := orchestrator.New(cfg.While, logger, workflow, runner)

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
	if err := orc.TearDown(); err != nil {
		logger.Fatalln("Failed teardown:", err)
	}
}
