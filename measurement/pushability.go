package measurement

import (
	"bytes"
	"fmt"
	"log"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type pushability struct {
	name                                 string
	PushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	Runner                               cmdRunner.CmdRunner
	RunnerOutBuf                         *bytes.Buffer
	RunnerErrBuf                         *bytes.Buffer
}

func (p *pushability) Name() string {
	return p.name
}

func (p *pushability) PerformMeasurement(logger *log.Logger, rs ResultSet) {
	defer p.RunnerOutBuf.Reset()
	defer p.RunnerErrBuf.Reset()

	if err := p.Runner.RunInSequence(p.PushAndDeleteAppCommandGeneratorFunc()...); err != nil {
		p.recordAndLogFailure(logger, err.Error(), p.RunnerOutBuf.String(), p.RunnerErrBuf.String(), rs)
		return
	}

	rs.RecordSuccess()
}

func (p *pushability) recordAndLogFailure(logger *log.Logger, errString, cmdOut, cmdErr string, rs ResultSet) {
	rs.RecordFailure()
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

func (p *pushability) Summary(rs ResultSet) string {
	if rs.Failed() > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d attempts to push and delete an app failed", p.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d attempts to push and delete an app succeeded", p.name, rs.Total())
}
