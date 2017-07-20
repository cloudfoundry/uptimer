package measurement

import (
	"fmt"
	"log"
	"net/http"
)

type availability struct {
	name          string
	summaryPhrase string
	url           string
	client        *http.Client
}

func (a *availability) Name() string {
	return a.name
}

func (a *availability) SummaryPhrase() string {
	return a.summaryPhrase
}

func (a *availability) PerformMeasurement(logger *log.Logger) bool {
	res, err := a.client.Get(a.url)
	if err != nil {
		a.logFailure(logger, err.Error())
		return false
	} else if res.StatusCode != http.StatusOK {
		a.logFailure(logger, fmt.Sprintf("response had status %d", res.StatusCode))
		return false
	}

	return true
}

func (a *availability) logFailure(logger *log.Logger, msg string) {
	logger.Printf("\x1b[31mFAILURE(%s): %s\x1b[0m\n", a.name, msg)
}

func (a *availability) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}
