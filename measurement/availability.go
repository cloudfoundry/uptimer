package measurement

import (
	"fmt"
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

func (a *availability) PerformMeasurement() (string, bool) {
	res, err := a.client.Get(a.url)
	if err != nil {
		return a.fmtFailure(err.Error()), false
	} else if res.StatusCode != http.StatusOK {
		return a.fmtFailure(fmt.Sprintf("response had status %d", res.StatusCode)), false
	}

	return "", true
}

func (a *availability) fmtFailure(msg string) string {
	return fmt.Sprintf("\x1b[31mFAILURE(%s): %s\x1b[0m\n", a.name, msg)
}

func (a *availability) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}
