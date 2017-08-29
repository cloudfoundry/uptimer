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
	"strings"
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
	"github.com/cloudfoundry/uptimer/version"
)

func main() {
	logger := log.New(os.Stdout, "\n[UPTIMER] ", log.Ldate|log.Ltime|log.LUTC)

	configPath := flag.String("configFile", "", "Path to the config file")
	showVersion := flag.Bool("v", false, "Prints the version of uptimer and exits")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Version)
		os.Exit(0)
	}

	if *configPath == "" {
		logger.Println("Failed to load config: ", fmt.Errorf("'-configFile' flag required"))
		os.Exit(1)
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Println("Failed to load config: ", err)
		os.Exit(1)
	}

	performMeasurements := true

	logger.Println("Building included app...")
	appPath, err := compileIncludedApp()
	if err != nil {
		logger.Println("Failed to build included app: ", err)
		performMeasurements = false
	}
	logger.Println("Finished building included app")

	orcTmpDir, recentLogsTmpDir, streamingLogsTmpDir, pushTmpDir, err := createTmpDirs()
	if err != nil {
		logger.Println("Failed to create temp dir:", err)
		performMeasurements = false
	}

	bufferedRunner, runnerOutBuf, runnerErrBuf := createBufferedRunner()

	pushCmdGenerator := cfCmdGenerator.New(pushTmpDir)
	pushWorkflow, pushOrg := createWorkflow(cfg.CF, appPath)
	logger.Printf("Setting up push workflow with org %s...", pushOrg)
	if err := bufferedRunner.RunInSequence(pushWorkflow.Setup(pushCmdGenerator)...); err != nil {
		logBufferedRunnerFailure(logger, "push workflow setup", err, runnerOutBuf, runnerErrBuf)
		performMeasurements = false
	} else {
		logger.Println("Finished setting up push workflow")
	}

	orcCmdGenerator := cfCmdGenerator.New(orcTmpDir)
	orcWorkflow, orcOrg := createWorkflow(cfg.CF, appPath)
	measurements := createMeasurements(
		logger,
		orcWorkflow,
		pushWorkflow,
		cfCmdGenerator.New(recentLogsTmpDir),
		cfCmdGenerator.New(streamingLogsTmpDir),
		pushCmdGenerator,
	)

	orc := orchestrator.New(cfg.While, logger, orcWorkflow, cmdRunner.New(os.Stdout, os.Stderr, io.Copy), measurements)

	logger.Printf("Setting up main workflow with org %s...", orcOrg)
	if err := orc.Setup(bufferedRunner, orcCmdGenerator); err != nil {
		logBufferedRunnerFailure(logger, "main workflow setup", err, runnerOutBuf, runnerErrBuf)
		performMeasurements = false
	} else {
		logger.Println("Finished setting up main workflow")
	}

	exitCode, err := orc.Run(performMeasurements)
	if err != nil {
		logger.Println("Failed run:", err)
	}

	logger.Println("Tearing down...")
	tearDown(
		orc,
		orcCmdGenerator,
		logger,
		pushWorkflow,
		pushCmdGenerator,
		bufferedRunner,
		runnerOutBuf,
		runnerErrBuf,
	)
	logger.Println("Finished tearing down")

	os.Exit(exitCode)
}

func loadConfig(configPath string) (*config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func createTmpDirs() (string, string, string, string, error) {
	orcTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		return "", "", "", "", err
	}
	recentLogsTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		return "", "", "", "", err
	}
	streamingLogsTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		return "", "", "", "", err
	}
	pushTmpDir, err := ioutil.TempDir("", "uptimer")
	if err != nil {
		return "", "", "", "", err
	}

	return orcTmpDir, recentLogsTmpDir, streamingLogsTmpDir, pushTmpDir, nil
}

func compileIncludedApp() (string, error) {
	appPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/cloudfoundry/uptimer/app")

	buildCmd := exec.Command("go", "build")
	buildCmd.Dir = appPath
	buildCmd.Env = []string{"GOOS=linux", "GOARCH=amd64"}
	err := buildCmd.Run()

	return appPath, err
}

func createWorkflow(cfc *config.Cf, appPath string) (cfWorkflow.CfWorkflow, string) {
	workflowOrg := fmt.Sprintf("uptimer-org-%s", uuid.NewV4().String())

	return cfWorkflow.New(
			cfc,
			workflowOrg,
			fmt.Sprintf("uptimer-space-%s", uuid.NewV4().String()),
			fmt.Sprintf("uptimer-quota-%s", uuid.NewV4().String()),
			fmt.Sprintf("uptimer-app-%s", uuid.NewV4().String()),
			appPath,
		),
		workflowOrg
}

func createMeasurements(logger *log.Logger, orcWorkflow, pushWorkflow cfWorkflow.CfWorkflow, recentLogsCmdGenerator, streamingLogsCmdGenerator, pushCmdGenerator cfCmdGenerator.CfCmdGenerator) []measurement.Measurement {
	recentLogsBufferRunner, recentLogsRunnerOutBuf, recentLogsRunnerErrBuf := createBufferedRunner()
	recentLogsMeasurement := measurement.NewRecentLogs(
		func() []cmdStartWaiter.CmdStartWaiter {
			return orcWorkflow.RecentLogs(recentLogsCmdGenerator)
		},
		recentLogsBufferRunner,
		recentLogsRunnerOutBuf,
		recentLogsRunnerErrBuf,
		appLogValidator.New(),
	)

	streamLogsBufferRunner, streamLogsRunnerOutBuf, streamLogsRunnerErrBuf := createBufferedRunner()
	streamLogsMeasurement := measurement.NewStreamLogs(
		func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter) {
			ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
			return ctx, cancelFunc, orcWorkflow.StreamLogs(ctx, streamingLogsCmdGenerator)
		},
		streamLogsBufferRunner,
		streamLogsRunnerOutBuf,
		streamLogsRunnerErrBuf,
		appLogValidator.New(),
	)

	pushRunner, pushRunnerOutBuf, pushRunnerErrBuf := createBufferedRunner()
	pushabilityMeasurement := measurement.NewPushability(
		func() []cmdStartWaiter.CmdStartWaiter {
			return append(pushWorkflow.Push(pushCmdGenerator), pushWorkflow.Delete(pushCmdGenerator)...)
		},
		pushRunner,
		pushRunnerOutBuf,
		pushRunnerErrBuf,
	)

	availabilityMeasurement := measurement.NewAvailability(
		orcWorkflow.AppUrl(),
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	)

	authFailedRetryFunc := func(stdOut, stdErr string) bool {
		authFailedMessage := "Authentication has expired.  Please log back in to re-authenticate."
		return strings.Contains(stdOut, authFailedMessage) || strings.Contains(stdErr, authFailedMessage)
	}

	clock := clock.New()
	return []measurement.Measurement{
		measurement.NewPeriodic(
			logger,
			clock,
			time.Second,
			availabilityMeasurement,
			measurement.NewResultSet(),
			func(string, string) bool { return false },
		),
		measurement.NewPeriodic(
			logger,
			clock,
			time.Minute,
			pushabilityMeasurement,
			measurement.NewResultSet(),
			authFailedRetryFunc,
		),
		measurement.NewPeriodic(
			logger,
			clock,
			10*time.Second,
			recentLogsMeasurement,
			measurement.NewResultSet(),
			authFailedRetryFunc,
		),
		measurement.NewPeriodic(
			logger,
			clock,
			30*time.Second,
			streamLogsMeasurement,
			measurement.NewResultSet(),
			authFailedRetryFunc,
		),
	}
}

func createBufferedRunner() (cmdRunner.CmdRunner, *bytes.Buffer, *bytes.Buffer) {
	outBuf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	return cmdRunner.New(outBuf, errBuf, io.Copy), outBuf, errBuf
}

func logBufferedRunnerFailure(
	logger *log.Logger,
	whatFailed string,
	err error,
	outBuf, errBuf *bytes.Buffer,
) {
	logger.Printf(
		"Failed %s: %v\nstdout:\n%s\nstderr:\n%s\n",
		whatFailed,
		err,
		outBuf.String(),
		errBuf.String(),
	)
	outBuf.Reset()
	errBuf.Reset()
}

func tearDown(
	orc orchestrator.Orchestrator,
	orcCmdGenerator cfCmdGenerator.CfCmdGenerator,
	logger *log.Logger,
	pushWorkflow cfWorkflow.CfWorkflow,
	pushCmdGenerator cfCmdGenerator.CfCmdGenerator,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
) {
	if err := orc.TearDown(runner, orcCmdGenerator); err != nil {
		logBufferedRunnerFailure(logger, "main teardown", err, runnerOutBuf, runnerErrBuf)
	}

	if err := runner.RunInSequence(pushWorkflow.TearDown(pushCmdGenerator)...); err != nil {
		logBufferedRunnerFailure(logger, "push workflow teardown", err, runnerOutBuf, runnerErrBuf)
	}
}
