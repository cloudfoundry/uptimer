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

//go:generate counterfeiter . ResultSet
type ResultSet interface {
	Successful() int
	Failed() int
	Total() int
}

type resultSet struct {
	successful int
	failed     int
}

func (rs *resultSet) Successful() int {
	return rs.successful
}

func (rs *resultSet) Failed() int {
	return rs.failed
}

func (rs *resultSet) Total() int {
	return rs.successful + rs.failed
}

func NewAvailability(logger *log.Logger, url string, frequency time.Duration, clock clock.Clock, client *http.Client) Measurement {
	return &availability{
		name:      "HTTP availability",
		logger:    logger,
		URL:       url,
		Frequency: frequency,
		Clock:     clock,
		Client:    client,
		resultSet: &resultSet{},
		stopChan:  make(chan int),
	}
}

func NewRecentLogs(
	frequency time.Duration,
	clock clock.Clock,
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
	runnerBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) Measurement {
	return &recentLogs{
		name: "Recent logs fetching",
		RecentLogsCommandGeneratorFunc: recentLogsCommandGeneratorFunc,
		Runner:          runner,
		RunnerBuf:       runnerBuf,
		appLogValidator: appLogValidator,
		Frequency:       frequency,
		Clock:           clock,
		resultSet:       &resultSet{},
		stopChan:        make(chan int),
	}
}

func NewPushability(
	logger *log.Logger,
	frequency time.Duration,
	clock clock.Clock,
	pushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter,
	runner cmdRunner.CmdRunner,
) Measurement {
	return &pushability{
		name:   "App pushability",
		logger: logger,
		PushAndDeleteAppCommandGeneratorFunc: pushAndDeleteAppCommandGeneratorFunc,
		Runner:    runner,
		Frequency: frequency,
		Clock:     clock,
		resultSet: &resultSet{},
		stopChan:  make(chan int),
	}
}

func NewStreamLogs(
	frequency time.Duration,
	clock clock.Clock,
	streamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter),
	runner cmdRunner.CmdRunner,
	runnerBuf *bytes.Buffer,
	appLogValidator appLogValidator.AppLogValidator,
) Measurement {
	return &streamLogs{
		name: "Streaming logs",
		StreamLogsCommandGeneratorFunc: streamLogsCommandGeneratorFunc,
		Runner:          runner,
		RunnerBuf:       runnerBuf,
		appLogValidator: appLogValidator,
		Frequency:       frequency,
		Clock:           clock,
		resultSet:       &resultSet{},
		stopChan:        make(chan int),
	}
}
