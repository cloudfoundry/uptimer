package measurement

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type streamLogs struct {
	name                           string
	summaryPhrase                  string
	streamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter)
	runner                         cmdRunner.CmdRunner
	runnerOutBuf                   *bytes.Buffer
	runnerErrBuf                   *bytes.Buffer
	appLogValidator                appLogValidator.AppLogValidator
}

func (s *streamLogs) Name() string {
	return s.name
}

func (s *streamLogs) SummaryPhrase() string {
	return s.summaryPhrase
}

func (s *streamLogs) PerformMeasurement() (string, bool) {
	defer s.runnerOutBuf.Reset()
	defer s.runnerErrBuf.Reset()

	ctx, cancelFunc, cmds := s.streamLogsCommandGeneratorFunc()
	defer cancelFunc()

	if err := s.runner.RunInSequenceWithContext(ctx, cmds...); err != nil {
		return s.fmtFailure(
				err.Error(),
				s.runnerOutBuf.String(),
				s.runnerErrBuf.String(),
			),
			false
	}

	logIsNewer, err := s.appLogValidator.IsNewer(s.runnerOutBuf.String())
	if err != nil {
		return s.fmtFailure(
				fmt.Sprintf("App log validation failed with: %s", err.Error()),
				s.runnerOutBuf.String(),
				s.runnerErrBuf.String(),
			),
			false

	} else if !logIsNewer {
		return s.fmtFailure(
				"App log fetched was not newer than previous app log fetched",
				s.runnerOutBuf.String(),
				s.runnerErrBuf.String(),
			),
			false
	}

	return "", true
}

func (s *streamLogs) fmtFailure(errString, cmdOut, cmdErr string) string {
	return fmt.Sprintf(
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
