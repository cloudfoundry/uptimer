package measurement

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type streamLogs struct {
	name                           string
	StreamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter)
	Runner                         cmdRunner.CmdRunner
	RunnerOutBuf                   *bytes.Buffer
	RunnerErrBuf                   *bytes.Buffer
	appLogValidator                appLogValidator.AppLogValidator
}

func (s *streamLogs) Name() string {
	return s.name
}

func (s *streamLogs) PerformMeasurement(logger *log.Logger, rs ResultSet) {
	defer s.RunnerOutBuf.Reset()
	defer s.RunnerErrBuf.Reset()

	ctx, cancelFunc, cmds := s.StreamLogsCommandGeneratorFunc()
	defer cancelFunc()

	if err := s.Runner.RunInSequenceWithContext(ctx, cmds...); err != nil {
		s.recordAndLogFailure(logger, err.Error(), s.RunnerOutBuf.String(), s.RunnerErrBuf.String(), rs)
		return
	}

	logIsNewer, err := s.appLogValidator.IsNewer(s.RunnerOutBuf.String())
	if err == nil && logIsNewer {
		rs.RecordSuccess()
		return
	}

	if err != nil {
		s.recordAndLogFailure(logger, fmt.Sprintf("App log validation failed with: %s", err.Error()), s.RunnerOutBuf.String(), s.RunnerErrBuf.String(), rs)

	} else if !logIsNewer {
		s.recordAndLogFailure(logger, "App log fetched was not newer than previous app log fetched", s.RunnerOutBuf.String(), s.RunnerErrBuf.String(), rs)
	}
}

func (s *streamLogs) recordAndLogFailure(logger *log.Logger, errString, cmdOut, cmdErr string, rs ResultSet) {
	rs.RecordFailure()
	logger.Printf(
		"\x1b[31mFAILURE(%s): %s\x1b[0m\nstdout:\n%s\nstderr:\n%s\n\n",
		s.name,
		errString,
		cmdOut,
		cmdErr,
	)
}

func (s *streamLogs) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}

func (s *streamLogs) Summary(rs ResultSet) string {
	if rs.Failed() > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d attempts to stream logs failed", s.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d attempts to stream logs succeeded", s.name, rs.Total())
}
