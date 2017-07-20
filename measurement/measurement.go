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

func NewPeriodic(logger *log.Logger, clock clock.Clock, freq time.Duration, baseMeasurement BaseMeasurement, resultSet ResultSet) Measurement {
	return &periodic{
		logger:          logger,
		clock:           clock,
		freq:            freq,
		baseMeasurement: baseMeasurement,

		stopChan:  make(chan int),
		resultSet: resultSet,
	}
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
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) BaseMeasurement {
	return &recentLogs{
		name: "Recent logs fetching",
		RecentLogsCommandGeneratorFunc: recentLogsCommandGeneratorFunc,
		Runner:          runner,
		RunnerOutBuf:    runnerOutBuf,
		RunnerErrBuf:    runnerErrBuf,
		appLogValidator: appLogValidator,
	}
}

func NewStreamLogs(
	streamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter),
	runner cmdRunner.CmdRunner,
	runnerOutBuf *bytes.Buffer,
	runnerErrBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) BaseMeasurement {
	return &streamLogs{
		name: "Streaming logs",
		StreamLogsCommandGeneratorFunc: streamLogsCommandGeneratorFunc,
		Runner:          runner,
		RunnerOutBuf:    runnerOutBuf,
		RunnerErrBuf:    runnerErrBuf,
		appLogValidator: appLogValidator,
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
