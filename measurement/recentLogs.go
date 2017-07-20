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
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	runner                         cmdRunner.CmdRunner
	runnerOutBuf                   *bytes.Buffer
	runnerErrBuf                   *bytes.Buffer
	appLogValidator                appLogValidator.AppLogValidator
}

func (r *recentLogs) Name() string {
	return r.name
}

func (r *recentLogs) PerformMeasurement(logger *log.Logger, rs ResultSet) {
	defer r.runnerOutBuf.Reset()
	defer r.runnerErrBuf.Reset()

	if err := r.runner.RunInSequence(r.recentLogsCommandGeneratorFunc()...); err != nil {
		r.recordAndLogFailure(logger, err.Error(), r.runnerOutBuf.String(), r.runnerErrBuf.String(), rs)
		return
	}

	logIsNewer, err := r.appLogValidator.IsNewer(r.runnerOutBuf.String())
	if err == nil && logIsNewer {
		rs.RecordSuccess()
		return
	}

	if err != nil {
		r.recordAndLogFailure(logger, fmt.Sprintf("App log validation failed with: %s", err.Error()), r.runnerOutBuf.String(), r.runnerErrBuf.String(), rs)

	} else if !logIsNewer {
		r.recordAndLogFailure(logger, "App log fetched was not newer than previous app log fetched", r.runnerOutBuf.String(), r.runnerErrBuf.String(), rs)
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
