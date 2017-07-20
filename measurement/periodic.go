package measurement

import (
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
		p.baseMeasurement.PerformMeasurement(p.logger, p.resultSet)
		for {
			select {
			case <-ticker.C:
				p.baseMeasurement.PerformMeasurement(p.logger, p.resultSet)
			case <-p.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
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
	return p.baseMeasurement.Summary(p.resultSet)
}
