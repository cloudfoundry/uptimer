package measurement

import (
	"bytes"
	"strings"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type statsAvailability struct {
	name                                  string
	summaryPhrase                         string
	statsAvailabilityCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	runner                                cmdRunner.CmdRunner
	runnerOutBuf                          *bytes.Buffer
	runnerErrBuf                          *bytes.Buffer
}

func (s *statsAvailability) Name() string {
	return s.name
}

func (s *statsAvailability) SummaryPhrase() string {
	return s.summaryPhrase
}

func (s *statsAvailability) PerformMeasurement() (string, string, string, bool) {
	defer s.runnerOutBuf.Reset()
	defer s.runnerErrBuf.Reset()

	if err := s.runner.RunInSequence(s.statsAvailabilityCommandGeneratorFunc()...); err != nil {
		return err.Error(), s.runnerOutBuf.String(), s.runnerErrBuf.String(), false
	}

	if strings.Contains(s.runnerErrBuf.String(), "Stats server temporarily unavailable.") {
		return "Stats server was unavailable",
			s.runnerOutBuf.String(),
			s.runnerErrBuf.String(),
			false
	}

	return "", "", "", true
}
