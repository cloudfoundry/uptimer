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
	msg, ok := p.baseMeasurement.PerformMeasurement()
	if !ok {
		p.resultSet.RecordFailure()
		p.logger.Print(msg)
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
		return fmt.Sprintf(
			"FAILED(%s): %d of %d attempts to %s failed",
			p.baseMeasurement.Name(),
			p.resultSet.Failed(),
			p.resultSet.Total(),
			p.baseMeasurement.SummaryPhrase(),
		)
	}

	return fmt.Sprintf(
		"SUCCESS(%s): All %d attempts to %s succeeded",
		p.baseMeasurement.Name(),
		p.resultSet.Total(),
		p.baseMeasurement.SummaryPhrase(),
	)
}
