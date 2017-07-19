package measurement

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

type availability struct {
	name      string
	logger    *log.Logger
	URL       string
	Frequency time.Duration
	Clock     clock.Clock
	Client    *http.Client

	resultSet *resultSet
	stopChan  chan int
}

func (a *availability) Name() string {
	return a.name
}

func (a *availability) Start() error {
	ticker := a.Clock.Ticker(a.Frequency)
	go func() {
		a.performRequest()
		for {
			select {
			case <-ticker.C:
				a.performRequest()
			case <-a.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (a *availability) performRequest() {
	res, err := a.Client.Get(a.URL)
	if err != nil {
		a.recordAndLogFailure(err.Error())
	} else if res.StatusCode != http.StatusOK {
		a.recordAndLogFailure(fmt.Sprintf("response had status %d", res.StatusCode))
	} else {
		a.resultSet.successful++
	}
}

func (a *availability) recordAndLogFailure(msg string) {
	a.resultSet.failed++
	a.logger.Printf("\x1b[31mFAILURE(%s): %s\x1b[0m\n", a.name, msg)
}

func (a *availability) Stop() error {
	a.stopChan <- 0
	return nil
}

func (a *availability) Results() (ResultSet, error) {
	return a.resultSet, nil
}

func (a *availability) Failed() bool {
	return a.resultSet.failed > 0
}
func (a *availability) Summary() string {
	rs := a.resultSet
	if rs.failed > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d requests failed", a.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d requests succeeded", a.name, rs.Total())
}
