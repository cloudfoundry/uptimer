package measurement

import (
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

type Measurement interface {
	Start() error
	Stop() error
	Results() (ResultSet, error)
}

type ResultSet interface {
	Successful() int
	Failed() int
	Total() int
}

func NewAvailability(url string, frequency time.Duration, clock clock.Clock, client *http.Client) Measurement {
	return &availability{
		URL:       url,
		Frequency: frequency,
		Clock:     clock,
		Client:    client,
		resultSet: &availabilityResultSet{},
		stopChan:  make(chan int),
	}
}
