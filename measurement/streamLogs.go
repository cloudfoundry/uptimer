package measurement

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type streamLogs struct {
	name                           string
	logger                         *log.Logger
	StreamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter)
	Runner                         cmdRunner.CmdRunner
	RunnerOutBuf                   *bytes.Buffer
	RunnerErrBuf                   *bytes.Buffer
	Frequency                      time.Duration
	Clock                          clock.Clock
	appLogValidator                appLogValidator.AppLogValidator

	resultSet *resultSet
	stopChan  chan int
}

func (s *streamLogs) Name() string {
	return s.name
}

func (s *streamLogs) Start() error {
	ticker := s.Clock.Ticker(s.Frequency)
	go func() {
		s.streamLogs()
		for {
			select {
			case <-ticker.C:
				s.streamLogs()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (s *streamLogs) streamLogs() {
	ctx, cancelFunc, cmds := s.StreamLogsCommandGeneratorFunc()
	defer cancelFunc()
	defer s.RunnerOutBuf.Reset()
	defer s.RunnerErrBuf.Reset()
	if err := s.Runner.RunInSequenceWithContext(ctx, cmds...); err != nil {
		s.recordAndLogFailure(err.Error(), s.RunnerOutBuf.String(), s.RunnerErrBuf.String())
		return
	}

	logIsNewer, err := s.appLogValidator.IsNewer(s.RunnerOutBuf.String())
	if err == nil && logIsNewer {
		s.resultSet.successful++
		return
	}

	if err != nil {
		s.recordAndLogFailure(fmt.Sprintf("App log validation failed with: %s", err.Error()), s.RunnerOutBuf.String(), s.RunnerErrBuf.String())

	} else if !logIsNewer {
		s.recordAndLogFailure("App log fetched was not newer than previous app log fetched", s.RunnerOutBuf.String(), s.RunnerErrBuf.String())
	}
}

func (s *streamLogs) recordAndLogFailure(errString, cmdOut, cmdErr string) {
	s.resultSet.failed++
	s.logger.Printf(
		"\x1b[31mFAILURE(%s): %s\x1b[0m\nstdout:\n%s\nstderr:\n%s\n\n",
		s.name,
		errString,
		cmdOut,
		cmdErr,
	)
}

func (s *streamLogs) Stop() error {
	s.stopChan <- 0
	return nil
}

func (s *streamLogs) Results() (ResultSet, error) {
	return s.resultSet, nil
}

func (s *streamLogs) Failed() bool {
	return s.resultSet.failed > 0
}

func (s *streamLogs) Summary() string {
	rs := s.resultSet
	if rs.failed > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d attempts to stream logs failed", s.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d attempts to stream logs succeeded", s.name, rs.Total())
}
