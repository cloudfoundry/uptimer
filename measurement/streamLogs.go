package measurement

import (
	"bytes"
	"fmt"
	"time"

	"context"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type streamLogs struct {
	StreamLogsCommandGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter)
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
	if err := s.Runner.RunInSequenceWithContext(ctx, cmds...); err != nil {
		s.resultSet.failed++
		return
	}

	logIsNewer, err := s.appLogValidator.IsNewer(s.RunnerBuf.String())
	s.RunnerBuf.Reset()
	if err != nil || !logIsNewer {
		s.resultSet.failed++
		return
	}

	s.resultSet.successful++
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
