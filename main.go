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
	logger := log.New(os.Stdout, "[UPTIMER] ", log.Ldate|log.Ltime|log.LUTC)

	cfg, err := loadConfig()
	if err != nil {
		logger.Println("Failed to load config: ", err)
		os.Exit(1)
	}

	logger.Println("Building included app")
	appPath, err := compileIncludedApp()
	if err != nil {
		logger.Println("Failed to build included app: ", err)
		os.Exit(1)
	}
	logger.Println("Finished building included app")

	baseTmpDir, pushTmpDir, err := createTmpDirs()
	if err != nil {
		logger.Println("Failed to create temp dir:", err)
		os.Exit(1)
	}

	logger.Println("Setting up push workflow")
	pushWorkflow := createWorkflow(cfg.CF, cfCmdGenerator.New(pushTmpDir), appPath)
	discardRunner := cmdRunner.New(ioutil.Discard, ioutil.Discard, io.Copy)
	if err := discardRunner.RunInSequence(pushWorkflow.Setup()...); err != nil {
		logger.Println("Failed push workflow setup: ", err)
		if err := discardRunner.RunInSequence(pushWorkflow.TearDown()...); err != nil {
			logger.Println("Failed push workflow teardown: ", err)
		}
		os.Exit(1)
	}
	logger.Println("Finished setting up push workflow")

	orcWorkflow := createWorkflow(cfg.CF, cfCmdGenerator.New(baseTmpDir), appPath)
	stdOutAndErrRunner := cmdRunner.New(os.Stdout, os.Stderr, io.Copy)
	measurements := createMeasurements(logger, orcWorkflow, pushWorkflow)

	orc := orchestrator.New(cfg.While, logger, orcWorkflow, stdOutAndErrRunner, measurements)

	logger.Println("Setting up")
	if err := orc.Setup(); err != nil {
		logger.Println("Failed setup:", err)
		tearDownAndExit(orc, logger, pushWorkflow, stdOutAndErrRunner, 1)
	}

	exitCode, err := orc.Run()
	if err != nil {
		logger.Println("Failed run:", err)
		tearDownAndExit(orc, logger, pushWorkflow, stdOutAndErrRunner, exitCode)
	}

	tearDownAndExit(orc, logger, pushWorkflow, stdOutAndErrRunner, exitCode)
}

func loadConfig() (*config.Config, error) {
	configPath := flag.String("configFile", "", "Path to the config file")
	flag.Parse()
	if *configPath == "" {
		return nil, fmt.Errorf("'-configFile' flag required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func createTmpDirs() (string, string, error) {
	baseTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		return "", "", err
	}
	pushTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		return "", "", err
	}

	return baseTmpDir, pushTmpDir, nil
}

func compileIncludedApp() (string, error) {
	appPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/cloudfoundry/uptimer/app")

	buildCmd := exec.Command("go", "build")
	buildCmd.Dir = appPath
	buildCmd.Env = []string{"GOOS=linux", "GOARCH=amd64"}
	err := buildCmd.Run()

	return appPath, err
}

func createWorkflow(cfc *config.CfConfig, cg cfCmdGenerator.CfCmdGenerator, appPath string) cfWorkflow.CfWorkflow {
	return cfWorkflow.New(
		cfc,
		cg,
		fmt.Sprintf("uptimer-org-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-space-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-quota-%s", uuid.NewV4().String()),
		fmt.Sprintf("uptimer-app-%s", uuid.NewV4().String()),
		appPath,
	)
}

func createMeasurements(logger *log.Logger, orcWorkflow, pushWorkflow cfWorkflow.CfWorkflow) []measurement.Measurement {
	recentLogsBufferRunner, recentLogsRunnerOutBuf, recentLogsRunnerErrBuf := createBufferedRunner()
	streamLogsBufferRunner, streamLogsRunnerOutBuf, streamLogsRunnerErrBuf := createBufferedRunner()
	pushRunner, pushRunnerOutBuf, pushRunnerErrBuf := createBufferedRunner()

	return []measurement.Measurement{
		measurement.NewAvailability(
			logger,
			orcWorkflow.AppUrl(),
			time.Second,
			clock.New(),
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		),
		measurement.NewRecentLogs(
			logger,
			10*time.Second,
			clock.New(),
			orcWorkflow.RecentLogs,
			recentLogsBufferRunner,
			recentLogsRunnerOutBuf,
			recentLogsRunnerErrBuf,
			appLogValidator.New(),
		),
		measurement.NewStreamLogs(
			logger,
			30*time.Second,
			clock.New(),
			func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
				return ctx, cancelFunc, orcWorkflow.StreamLogs(ctx)
			},
			streamLogsBufferRunner,
			streamLogsRunnerOutBuf,
			streamLogsRunnerErrBuf,
			appLogValidator.New(),
		),
		measurement.NewPushability(
			logger,
			time.Minute,
			clock.New(),
			func() []cmdStartWaiter.CmdStartWaiter {
				return append(pushWorkflow.Push(), pushWorkflow.Delete()...)
			},
			pushRunner,
			pushRunnerOutBuf,
			pushRunnerErrBuf,
		),
	}
}

func createBufferedRunner() (cmdRunner.CmdRunner, *bytes.Buffer, *bytes.Buffer) {
	outBuf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	return cmdRunner.New(outBuf, errBuf, io.Copy), outBuf, errBuf
}

func tearDownAndExit(orc orchestrator.Orchestrator, logger *log.Logger, pushWorkflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner, exitCode int) {
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
