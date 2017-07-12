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

func NewAvailability(url string, frequency time.Duration, clock clock.Clock, client *http.Client) Measurement {
	return &availability{
		name:      "HTTP availability",
		URL:       url,
		Frequency: frequency,
		Clock:     clock,
		Client:    client,
		resultSet: &availabilityResultSet{},
		stopChan:  make(chan int),
	}
}
