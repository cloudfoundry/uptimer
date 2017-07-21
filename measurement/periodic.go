package measurement

import (
	"fmt"
	"log"
	"time"

	"github.com/benbjohnson/clock"
)

type periodic struct {
	logger          *log.Logger
	clock           clock.Clock
	freq            time.Duration
	baseMeasurement BaseMeasurement

	resultSet ResultSet
	stopChan  chan int
}

func (p *periodic) Name() string {
	return p.baseMeasurement.Name()
}

func (p *periodic) Start() {
	ticker := p.clock.Ticker(p.freq)
	go func() {
		p.performMeasurement()
		for {
			select {
			case <-ticker.C:
				p.performMeasurement()
			case <-p.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *periodic) performMeasurement() {
	if msg, ok := p.baseMeasurement.PerformMeasurement(); !ok {
		sslf, lf := p.resultSet.SuccessesSinceLastFailure()
		p.resultSet.RecordFailure()

		p.logger.Println(msg)
		if sslf > 0 {
			p.logger.Printf("%d successes since last failure (at %s)\n", sslf, lf.Format("2006/01/02 15:04:05"))
		}
		return
	}

	p.resultSet.RecordSuccess()
}

func (p *periodic) Results() ResultSet {
	return p.resultSet
}

func (p *periodic) Stop() {
	p.stopChan <- 0
}

func (p *periodic) Failed() bool {
	return p.baseMeasurement.Failed(p.resultSet)
}

func (p *periodic) Summary() string {
	if p.Failed() {
		failMsg := fmt.Sprintf(
			"FAILED(%s): %d of %d attempts to %s failed",
			p.baseMeasurement.Name(),
			p.resultSet.Failed(),
			p.resultSet.Total(),
			p.baseMeasurement.SummaryPhrase(),
		)

		if sslf, _ := p.resultSet.SuccessesSinceLastFailure(); sslf > 0 {
			failMsg = fmt.Sprintf("%s (the last %d succeeded)", failMsg, sslf)
		}

		return failMsg
	}

	return fmt.Sprintf(
		"SUCCESS(%s): All %d attempts to %s succeeded",
		p.baseMeasurement.Name(),
		p.resultSet.Total(),
		p.baseMeasurement.SummaryPhrase(),
	)
}
