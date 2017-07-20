package measurement

import (
	"fmt"
	"log"
	"net/http"
)

type availability struct {
	name   string
	url    string
	client *http.Client
}

func (a *availability) Name() string {
	return a.name
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

func (a *availability) Summary(rs ResultSet) string {
	if rs.Failed() > 0 {
		return fmt.Sprintf("FAILED(%s): %d of %d requests failed", a.name, rs.Failed(), rs.Total())
	}

	return fmt.Sprintf("SUCCESS(%s): All %d requests succeeded", a.name, rs.Total())
}
