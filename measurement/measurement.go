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
	Start() error
	Stop() error
	Results() (ResultSet, error)
	Failed() bool
	Summary() string
}

//go:generate counterfeiter . BaseMeasurement
type BaseMeasurement interface {
	Name() string
	PerformMeasurement(*log.Logger, ResultSet)
	Failed(rs ResultSet) bool
	Summary(rs ResultSet) string
}

func NewAvailability(url string, client *http.Client) BaseMeasurement {
	return &availability{
		name:   "HTTP availability",
		url:    url,
		client: client,
	}
}

func NewRecentLogs(
	logger *log.Logger,
	frequency time.Duration,
	clock clock.Clock,
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) Measurement {
	return &recentLogs{
		name:   "Recent logs fetching",
		logger: logger,
		RecentLogsCommandGeneratorFunc: recentLogsCommandGeneratorFunc,
		Runner:          runner,
		RunnerOutBuf:    runnerOutBuf,
		RunnerErrBuf:    runnerErrBuf,
		appLogValidator: appLogValidator,
		Frequency:       frequency,
		Clock:           clock,
		resultSet:       &resultSet{},
		stopChan:        make(chan int),
	}
}

func NewStreamLogs(
	logger *log.Logger,
	frequency time.Duration,
	clock clock.Clock,
	streamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter),
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) Measurement {
	return &streamLogs{
		name:   "Streaming logs",
		logger: logger,
		StreamLogsCommandGeneratorFunc: streamLogsCommandGeneratorFunc,
		Runner:          runner,
		RunnerOutBuf:    runnerOutBuf,
		RunnerErrBuf:    runnerErrBuf,
		appLogValidator: appLogValidator,
		Frequency:       frequency,
		Clock:           clock,
		resultSet:       &resultSet{},
		stopChan:        make(chan int),
	}
}

func NewPushability(
	pushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
) BaseMeasurement {
	return &pushability{
		name: "App pushability",
		PushAndDeleteAppCommandGeneratorFunc: pushAndDeleteAppCommandGeneratorFunc,
		Runner:       runner,
		RunnerOutBuf: runnerOutBuf,
		RunnerErrBuf: runnerErrBuf,
	}
}
