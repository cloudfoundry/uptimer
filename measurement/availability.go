package measurement

import (
	"bytes"
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

func (a *availability) PerformMeasurement() (string, string, string, bool) {
	res, err := a.client.Get(a.url)
	if err != nil {
		return err.Error(), "", "", false
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(res.Body)
		if err != nil {
			return err.Error(), "", "", false
		}
		return fmt.Sprintf("response had status %d; %s; %s", res.StatusCode, res.Status, buf.String()), "", "", false
	}

	return "", "", "", true
}
