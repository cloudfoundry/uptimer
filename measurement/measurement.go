package measurement

import (
	"net/http"
	"time"

	"bytes"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type Measurement interface {
	Name() string
	Start() error
	Stop() error
	Results() (ResultSet, error)
	Failed() bool
	Summary() string
}

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

func NewAvailability(url string, frequency time.Duration, clock clock.Clock, client *http.Client) Measurement {
	return &availability{
		name:      "HTTP availability",
		URL:       url,
		Frequency: frequency,
		Clock:     clock,
		Client:    client,
		resultSet: &resultSet{},
		stopChan:  make(chan int),
	}
}

func NewRecentLogs(frequency time.Duration, clock clock.Clock, recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter, runner cmdRunner.CmdRunner, logBuf *bytes.Buffer, appLogValidator appLogValidator.AppLogValidator) Measurement {
	return &recentLogs{
		name: "Recent logs fetching",
		RecentLogsCommandGeneratorFunc: recentLogsCommandGeneratorFunc,
		Runner:          runner,
		LogBuf:          logBuf,
		appLogValidator: appLogValidator,
		Frequency:       frequency,
		Clock:           clock,
		resultSet:       &resultSet{},
		stopChan:        make(chan int),
	}
}

func NewPushability(frequency time.Duration, clock clock.Clock, pushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter, runner cmdRunner.CmdRunner) Measurement {
	return &pushability{
		name: "App pushability",
		PushAndDeleteAppCommandGeneratorFunc: pushAndDeleteAppCommandGeneratorFunc,
		Runner:    runner,
		Frequency: frequency,
		Clock:     clock,
		resultSet: &resultSet{},
		stopChan:  make(chan int),
	}
}
