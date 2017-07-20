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

func (a *availability) PerformMeasurement(logger *log.Logger, rs ResultSet) {
	res, err := a.client.Get(a.url)
	if err != nil {
		a.recordAndLogFailure(logger, err.Error(), rs)
	} else if res.StatusCode != http.StatusOK {
		a.recordAndLogFailure(logger, fmt.Sprintf("response had status %d", res.StatusCode), rs)
	} else {
		rs.RecordSuccess()
	}
}

func (a *availability) recordAndLogFailure(logger *log.Logger, msg string, rs ResultSet) {
	rs.RecordFailure()
	logger.Printf("\x1b[31mFAILURE(%s): %s\x1b[0m\n", a.name, msg)
}

func (a *availability) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}
