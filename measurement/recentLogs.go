package measurement

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type recentLogs struct {
	name                           string
	logger                         *log.Logger
	RecentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	Runner                         cmdRunner.CmdRunner
	RunnerOutBuf                   *bytes.Buffer
	RunnerErrBuf                   *bytes.Buffer
	Frequency                      time.Duration
	Clock                          clock.Clock
	appLogValidator                appLogValidator.AppLogValidator

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
	defer r.RunnerOutBuf.Reset()
	defer r.RunnerErrBuf.Reset()
	if err := r.Runner.RunInSequence(r.RecentLogsCommandGeneratorFunc()...); err != nil {
		r.recordAndLogFailure(err.Error(), r.RunnerOutBuf.String(), r.RunnerErrBuf.String())
		return
	}

	logIsNewer, err := r.appLogValidator.IsNewer(r.RunnerOutBuf.String())
	if err == nil && logIsNewer {
		r.resultSet.successful++
		return
	}

	if err != nil {
		r.recordAndLogFailure(fmt.Sprintf("App log validation failed with: %s", err.Error()), r.RunnerOutBuf.String(), r.RunnerErrBuf.String())

	} else if !logIsNewer {
		r.recordAndLogFailure("App log fetched was not newer than previous app log fetched", r.RunnerOutBuf.String(), r.RunnerErrBuf.String())
	}
}

func (r *recentLogs) recordAndLogFailure(errString, cmdOut, cmdErr string) {
	r.resultSet.failed++
	r.logger.Printf(
		"\x1b[31mFAILURE(%s): %s\x1b[0m\nstdout:\n%s\nstderr:\n%s\n\n",
		r.name,
		errString,
		cmdOut,
		cmdErr,
	)
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
