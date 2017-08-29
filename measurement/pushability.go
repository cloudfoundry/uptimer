package measurement

import (
	"bytes"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type pushability struct {
	name                                 string
	summaryPhrase                        string
	pushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	runner                               cmdRunner.CmdRunner
	runnerOutBuf                         *bytes.Buffer
	runnerErrBuf                         *bytes.Buffer
}

func (p *pushability) Name() string {
	return p.name
}

func (p *pushability) SummaryPhrase() string {
	return p.summaryPhrase
}

func (p *pushability) PerformMeasurement() (string, string, string, bool) {
	defer p.runnerOutBuf.Reset()
	defer p.runnerErrBuf.Reset()

	if err := p.runner.RunInSequence(p.pushAndDeleteAppCommandGeneratorFunc()...); err != nil {
		return err.Error(), p.runnerOutBuf.String(), p.runnerErrBuf.String(), false
	}

	return "", "", "", true
}
