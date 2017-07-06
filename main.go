package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
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

	runner := cmdRunner.New(os.Stdout, os.Stderr, io.Copy)
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	logger.Printf("Running command: `%s %s`\n", cfg.Command, strings.Join(cfg.CommandArgs, " "))
	cmd := exec.Command(cfg.Command, cfg.CommandArgs...)
	if err := runner.Run(cmd); err != nil {
		logger.Fatalln("Failed running command:", err)
	}
	logger.Println("Finished running command")
}
