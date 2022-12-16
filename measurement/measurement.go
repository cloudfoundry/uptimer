package measurement

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

//go:generate counterfeiter . Measurement
type Measurement interface {
	Name() string
	Start()
	Stop()
	Results() ResultSet
	Failed() bool
	Summary() string
	SummaryData() Summary
}

type ShouldRetryFunc func(stdOut, stdErr string) bool

func NewPeriodicWithoutMeasuringImmediately(
	logger *log.Logger,
	clock clock.Clock,
	freq time.Duration,
	baseMeasurement BaseMeasurement,
	resultSet ResultSet,
	allowedFailures int,
	shouldRetryFunc ShouldRetryFunc,
) Measurement {
	return &periodic{
		logger:             logger,
		clock:              clock,
		freq:               freq,
		baseMeasurement:    baseMeasurement,
		shouldRetryFunc:    shouldRetryFunc,
		allowedFailures:    allowedFailures,
		measureImmediately: false,

		stopChan:  make(chan int, 1),
		resultSet: resultSet,
	}
}

func NewPeriodic(
	logger *log.Logger,
	clock clock.Clock,
	freq time.Duration,
	baseMeasurement BaseMeasurement,
	resultSet ResultSet,
	allowedFailures int,
	shouldRetryFunc ShouldRetryFunc,
) Measurement {
	return &periodic{
		logger:             logger,
		clock:              clock,
		freq:               freq,
		baseMeasurement:    baseMeasurement,
		shouldRetryFunc:    shouldRetryFunc,
		allowedFailures:    allowedFailures,
		measureImmediately: true,

		stopChan:  make(chan int, 1),
		resultSet: resultSet,
	}
}

//go:generate counterfeiter . BaseMeasurement
type BaseMeasurement interface {
	Name() string
	PerformMeasurement() (string, string, string, bool)
	SummaryPhrase() string
}

func NewHTTPAvailability(url string, client *http.Client) BaseMeasurement {
	return &availability{
		name:          "HTTP availability",
		summaryPhrase: "perform get requests",
		url:           url,
		client:        client,
	}
}

func NewRecentLogs(
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) BaseMeasurement {
	return &recentLogs{
		name:                           "Recent logs",
		summaryPhrase:                  "fetch recent logs",
		recentLogsCommandGeneratorFunc: recentLogsCommandGeneratorFunc,
		runner:                         runner,
		runnerOutBuf:                   runnerOutBuf,
		runnerErrBuf:                   runnerErrBuf,
		appLogValidator:                appLogValidator,
	}
}

func NewSyslogDrain(
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) BaseMeasurement {
	return &recentLogs{
		name:                           "App syslog availability",
		summaryPhrase:                  "check application syslogs",
		recentLogsCommandGeneratorFunc: recentLogsCommandGeneratorFunc,
		runner:          runner,
		runnerOutBuf:    runnerOutBuf,
		runnerErrBuf:    runnerErrBuf,
		appLogValidator: appLogValidator,
	}
}

func NewStreamingLogs(
	streamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter),
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) BaseMeasurement {
	return &streamLogs{
		name:                           "Streaming logs",
		summaryPhrase:                  "stream logs",
		streamLogsCommandGeneratorFunc: streamLogsCommandGeneratorFunc,
		runner:          runner,
		runnerOutBuf:    runnerOutBuf,
		runnerErrBuf:    runnerErrBuf,
		appLogValidator: appLogValidator,
	}
}

func NewAppPushability(
	pushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
) BaseMeasurement {
	return &pushability{
		name:                                 "App pushability",
		summaryPhrase:                        "push and delete an app",
		pushAndDeleteAppCommandGeneratorFunc: pushAndDeleteAppCommandGeneratorFunc,
		runner:       runner,
		runnerOutBuf: runnerOutBuf,
		runnerErrBuf: runnerErrBuf,
	}
}
