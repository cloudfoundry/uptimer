package measurement

import (
	"fmt"
	"log"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type pushability struct {
	name                                 string
	logger                               *log.Logger
	PushAndDeleteAppCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	Runner                               cmdRunner.CmdRunner
	Frequency                            time.Duration
	Clock                                clock.Clock
	resultSet                            *resultSet
	stopChan                             chan int
}

func (p *pushability) Name() string {
	return p.name
}

func (p *pushability) Start() error {
	ticker := p.Clock.Ticker(p.Frequency)
	go func() {
		p.pushIt()
		for {
			select {
			case <-ticker.C:
				p.pushIt()
			case <-p.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (p *pushability) pushIt() {
	if err := p.Runner.RunInSequence(p.PushAndDeleteAppCommandGeneratorFunc()...); err != nil {
		p.recordAndLogFailure(err.Error())
		return
	}

	p.resultSet.successful++
}

func (p *pushability) recordAndLogFailure(msg string) {
	p.resultSet.failed++
	p.logger.Printf("\x1b[31mFAILURE(%s): %s\x1b[0m\n", p.name, msg)
}

func (p *pushability) Stop() error {
	p.stopChan <- 0
	return nil
}

func (p *pushability) Results() (ResultSet, error) {
	return p.resultSet, nil
}

func (p *pushability) Failed() bool {
	return p.resultSet.failed > 0
}
func (p *pushability) Summary() string {
	rs := p.resultSet
	if rs.failed > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d attempts to push and delete an app failed", p.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d attempts to push and delete an app succeeded", p.name, rs.Total())
}
