package measurement

import (
	"bytes"
	"log"

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

func (p *pushability) PerformMeasurement(logger *log.Logger) bool {
	defer p.runnerOutBuf.Reset()
	defer p.runnerErrBuf.Reset()

	if err := p.runner.RunInSequence(p.pushAndDeleteAppCommandGeneratorFunc()...); err != nil {
		p.logFailure(logger, err.Error(), p.runnerOutBuf.String(), p.runnerErrBuf.String())
		return false
	}

	return true
}

func (p *pushability) logFailure(logger *log.Logger, errString, cmdOut, cmdErr string) {
	logger.Printf(
		"\x1b[31mFAILURE(%s): %s\x1b[0m\nstdout:\n%s\nstderr:\n%s\n\n",
		p.name,
		errString,
		cmdOut,
		cmdErr,
	)
}

func (p *pushability) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}
