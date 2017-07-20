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

func (s *streamLogs) PerformMeasurement(logger *log.Logger) bool {
	defer s.runnerOutBuf.Reset()
	defer s.runnerErrBuf.Reset()

	ctx, cancelFunc, cmds := s.streamLogsCommandGeneratorFunc()
	defer cancelFunc()

	if err := s.runner.RunInSequenceWithContext(ctx, cmds...); err != nil {
		s.logFailure(logger, err.Error(), s.runnerOutBuf.String(), s.runnerErrBuf.String())
		return false
	}

	logIsNewer, err := s.appLogValidator.IsNewer(s.runnerOutBuf.String())
	if err == nil && logIsNewer {
		return true
	}

	if err != nil {
		s.logFailure(logger, fmt.Sprintf("App log validation failed with: %s", err.Error()), s.runnerOutBuf.String(), s.runnerErrBuf.String())

	} else if !logIsNewer {
		s.logFailure(logger, "App log fetched was not newer than previous app log fetched", s.runnerOutBuf.String(), s.runnerErrBuf.String())
	}

	return false
}

func (s *streamLogs) logFailure(logger *log.Logger, errString, cmdOut, cmdErr string) {
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
