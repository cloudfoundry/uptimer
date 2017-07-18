package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/benbjohnson/clock"
	uuid "github.com/satori/go.uuid"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
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
	stdOutAndErrRunner := cmdRunner.New(os.Stdout, os.Stderr, io.Copy)

	baseTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		logger.Println("Failed to create base temp dir:", err)
		os.Exit(1)
	}
	baseCfCmdGenerator := cfCmdGenerator.New(baseTmpDir)

	appPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/cloudfoundry/uptimer/app")
	buildCmd := exec.Command("go", "build")
	buildCmd.Dir = appPath
	buildCmd.Env = []string{"GOOS=linux", "GOARCH=amd64"}
	logger.Println("Building included app")
	if err := buildCmd.Run(); err != nil {
		logger.Println("Failed to build included app: ", err)
		os.Exit(1)
	}
	logger.Println("Finished building included app")

	workflow := cfWorkflow.New(
		cfg.CF,
		baseCfCmdGenerator,
		fmt.Sprintf("uptimer-org-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-space-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-app-%s", uuid.NewV4().String()),
		appPath,
	)

	var recentLogsBuf = bytes.NewBuffer([]byte{})
	recentLogsBufferRunner := cmdRunner.New(recentLogsBuf, ioutil.Discard, io.Copy)
	var streamLogsBuf = bytes.NewBuffer([]byte{})
	streamLogsBufferRunner := cmdRunner.New(streamLogsBuf, ioutil.Discard, io.Copy)
	appLogValidator := appLogValidator.New()

	// We are copying values from the cfg object so that this workflow generates its own
	// org, space, and app names
	pushTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		logger.Println("Failed to create push temp dir:", err)
		os.Exit(1)
	}
	pushCfCmdGenerator := cfCmdGenerator.New(pushTmpDir)
	pushWorkflow := cfWorkflow.New(
		&config.CfConfig{
			API:           cfg.CF.API,
			AppDomain:     cfg.CF.AppDomain,
			AdminUser:     cfg.CF.AdminUser,
			AdminPassword: cfg.CF.AdminPassword,
		},
		pushCfCmdGenerator,
		fmt.Sprintf("uptimer-org-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-space-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-app-%s", uuid.NewV4().String()),
		appPath,
	)
	discardRunner := cmdRunner.New(ioutil.Discard, ioutil.Discard, io.Copy)
	if err := discardRunner.RunInSequence(pushWorkflow.Setup()...); err != nil {
		logger.Println("Failed push workflow setup: ", err)
		if err := discardRunner.RunInSequence(pushWorkflow.TearDown()...); err != nil {
			logger.Println("Failed push workflow teardown: ", err)
		}
		os.Exit(1)
	}

	measurements := []measurement.Measurement{
		measurement.NewAvailability(
			workflow.AppUrl(),
			time.Second,
			clock.New(),
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		),
		measurement.NewRecentLogs(
			10*time.Second,
			clock.New(),
			workflow.RecentLogs,
			recentLogsBufferRunner,
			recentLogsBuf,
			appLogValidator,
		),
		measurement.NewPushability(
			time.Minute,
			clock.New(),
			func() []cmdStartWaiter.CmdStartWaiter {
				return append(pushWorkflow.Push(), pushWorkflow.Delete()...)
			},
			discardRunner,
		),
		measurement.NewStreamLogs(
			30*time.Second,
			clock.New(),
			func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
				return ctx, cancelFunc, workflow.StreamLogs(ctx)
			},
			streamLogsBufferRunner,
			streamLogsBuf,
			appLogValidator,
		),
	}

	orc := orchestrator.New(cfg.While, logger, workflow, stdOutAndErrRunner, measurements)

	logger.Println("Setting up")
	if err := orc.Setup(); err != nil {
		logger.Println("Failed setup:", err)
		TearDownAndExit(orc, logger, pushWorkflow, stdOutAndErrRunner, 1)
	}

	exitCode, err := orc.Run()
	if err != nil {
		logger.Println("Failed run:", err)
		TearDownAndExit(orc, logger, pushWorkflow, stdOutAndErrRunner, exitCode)
	}

	TearDownAndExit(orc, logger, pushWorkflow, stdOutAndErrRunner, exitCode)
}

func TearDownAndExit(orc orchestrator.Orchestrator, logger *log.Logger, pushWorkflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner, exitCode int) {
	logger.Println("Tearing down")
	if err := orc.TearDown(); err != nil {
		logger.Fatalln("Failed main teardown:", err)
	}
	if err := runner.RunInSequence(pushWorkflow.TearDown()...); err != nil {
		logger.Println("Failed push workflow teardown: ", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}
