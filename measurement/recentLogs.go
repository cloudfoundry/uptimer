package measurement

import (
	"bytes"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type recentLogs struct {
	RecentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	Runner                         cmdRunner.CmdRunner
	RunnerBuf                      *bytes.Buffer
	Frequency                      time.Duration
	Clock                          clock.Clock
	appLogValidator                appLogValidator.AppLogValidator

	name      string
	resultSet *resultSet
	stopChan  chan int

	lastAppNumber int
}

func (r *recentLogs) Name() string {
	return r.name
}

func (r *recentLogs) Start() error {
	ticker := r.Clock.Ticker(r.Frequency)
	go func() {
		r.getRecentLogs()
		for {
			select {
			case <-ticker.C:
				r.getRecentLogs()
			case <-r.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (r *recentLogs) getRecentLogs() {
	if err := r.Runner.RunInSequence(r.RecentLogsCommandGeneratorFunc()...); err != nil {
		r.resultSet.failed++
		return
	}

	logIsNewer, err := r.appLogValidator.IsNewer(r.RunnerBuf.String())
	r.RunnerBuf.Reset()
	if err != nil || !logIsNewer {
		r.resultSet.failed++
		return
	}

	r.resultSet.successful++
}

func (r *recentLogs) Stop() error {
	r.stopChan <- 0
	return nil
}

func (r *recentLogs) Results() (ResultSet, error) {
	return r.resultSet, nil
}

func (r *recentLogs) Failed() bool {
	return r.resultSet.failed > 0
}
func (r *recentLogs) Summary() string {
	rs := r.resultSet
	if rs.failed > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d attempts to fetch recent logs failed", r.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d attempts to fetch recent logs succeeded", r.name, rs.Total())
}
