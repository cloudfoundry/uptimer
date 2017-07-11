package measurement

import (
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

type availabilityResultSet struct {
	successful int
	failed     int
}

func (rs *availabilityResultSet) Successful() int {
	return rs.successful
}

func (rs *availabilityResultSet) Failed() int {
	return rs.failed
}

func (rs *availabilityResultSet) Total() int {
	return rs.successful + rs.failed
}

type availability struct {
	URL       string
	Frequency time.Duration
	Clock     clock.Clock
	Client    *http.Client

	resultSet *availabilityResultSet
	stopChan  chan int
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
	if err != nil || res.StatusCode != http.StatusOK {
		a.resultSet.failed++
	} else {
		a.resultSet.successful++
	}
}

func (a *availability) Stop() error {
	a.stopChan <- 0
	return nil
}

func (a *availability) Results() (ResultSet, error) {
	return a.resultSet, nil
}
