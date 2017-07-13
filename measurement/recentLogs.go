package measurement

import (
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/cmdRunner"
)

type recentLogs struct {
	RecentLogsCommandGeneratorFunc func() []cmdRunner.CmdStartWaiter
	Runner                         cmdRunner.CmdRunner
	Frequency                      time.Duration
	Clock                          clock.Clock

	name      string
	resultSet *resultSet
	stopChan  chan int
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
	} else {
		r.resultSet.successful++
	}
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
