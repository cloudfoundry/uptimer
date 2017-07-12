package measurement

import (
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

type Measurement interface {
	Name() string
	Start() error
	Stop() error
	Results() (ResultSet, error)
	Failed() bool
	Summary() string
}

type ResultSet interface {
	Successful() int
	Failed() int
	Total() int
}

type resultSet struct {
	successful int
	failed     int
}

func (rs *resultSet) Successful() int {
	return rs.successful
}

func (rs *resultSet) Failed() int {
	return rs.failed
}

func (rs *resultSet) Total() int {
	return rs.successful + rs.failed
}

func NewAvailability(url string, frequency time.Duration, clock clock.Clock, client *http.Client) Measurement {
	return &availability{
		name:      "HTTP availability",
		URL:       url,
		Frequency: frequency,
		Clock:     clock,
		Client:    client,
		resultSet: &resultSet{},
		stopChan:  make(chan int),
	}
}
