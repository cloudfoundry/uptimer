package measurement

import (
	"bytes"
	"fmt"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type recentLogs struct {
	name                           string
	summaryPhrase                  string
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	runner                         cmdRunner.CmdRunner
	runnerOutBuf                   *bytes.Buffer
	runnerErrBuf                   *bytes.Buffer
	appLogValidator                appLogValidator.AppLogValidator
}

func (r *recentLogs) Name() string {
	return r.name
}

func (r *recentLogs) SummaryPhrase() string {
	return r.summaryPhrase
}

func (r *recentLogs) PerformMeasurement() (string, string, string, bool) {
	defer r.runnerOutBuf.Reset()
	defer r.runnerErrBuf.Reset()

	if err := r.runner.RunInSequence(r.recentLogsCommandGeneratorFunc()...); err != nil {
		return err.Error(), r.runnerOutBuf.String(), r.runnerErrBuf.String(), false
	}

	logIsNewer, err := r.appLogValidator.IsNewer(r.runnerOutBuf.String())
	if err != nil {
		return fmt.Sprintf("App log validation failed with: %s", err.Error()),
			r.runnerOutBuf.String(),
			r.runnerErrBuf.String(),
			false
	} else if !logIsNewer {
		return "App log fetched was not newer than previous app log fetched",
			r.runnerOutBuf.String(),
			r.runnerErrBuf.String(),
			false
	}

	return "", "", "", true
}
