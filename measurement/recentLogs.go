package measurement

import (
	"bytes"
	"fmt"
	"log"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type recentLogs struct {
	name                           string
	RecentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	Runner                         cmdRunner.CmdRunner
	RunnerOutBuf                   *bytes.Buffer
	RunnerErrBuf                   *bytes.Buffer
	appLogValidator                appLogValidator.AppLogValidator
}

func (r *recentLogs) Name() string {
	return r.name
}

func (r *recentLogs) PerformMeasurement(logger *log.Logger, rs ResultSet) {
	defer r.RunnerOutBuf.Reset()
	defer r.RunnerErrBuf.Reset()

	if err := r.Runner.RunInSequence(r.RecentLogsCommandGeneratorFunc()...); err != nil {
		r.recordAndLogFailure(logger, err.Error(), r.RunnerOutBuf.String(), r.RunnerErrBuf.String(), rs)
		return
	}

	logIsNewer, err := r.appLogValidator.IsNewer(r.RunnerOutBuf.String())
	if err == nil && logIsNewer {
		rs.RecordSuccess()
		return
	}

	if err != nil {
		r.recordAndLogFailure(logger, fmt.Sprintf("App log validation failed with: %s", err.Error()), r.RunnerOutBuf.String(), r.RunnerErrBuf.String(), rs)

	} else if !logIsNewer {
		r.recordAndLogFailure(logger, "App log fetched was not newer than previous app log fetched", r.RunnerOutBuf.String(), r.RunnerErrBuf.String(), rs)
	}
}

func (r *recentLogs) recordAndLogFailure(logger *log.Logger, errString, cmdOut, cmdErr string, rs ResultSet) {
	rs.RecordFailure()
	logger.Printf(
		"\x1b[31mFAILURE(%s): %s\x1b[0m\nstdout:\n%s\nstderr:\n%s\n\n",
		r.name,
		errString,
		cmdOut,
		cmdErr,
	)
}

func (r *recentLogs) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}
func (r *recentLogs) Summary(rs ResultSet) string {
	if rs.Failed() > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d attempts to fetch recent logs failed", r.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d attempts to fetch recent logs succeeded", r.name, rs.Total())
}
