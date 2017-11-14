package appLogValidator

import (
	"fmt"
	"strconv"
	"strings"
)

//go:generate counterfeiter . AppLogValidator
type AppLogValidator interface {
	IsNewer(log string) (bool, error)
}

type appLogValidator struct {
	lastNumber int
}

func New() AppLogValidator {
	return &appLogValidator{
		lastNumber: -1,
	}
}

func (v *appLogValidator) IsNewer(log string) (bool, error) {
	latestNumber, err := getLatestAppNumber(log)
	if err != nil {
		return false, err
	}

	if latestNumber <= v.lastNumber && v.lastNumber > -1 {
		return false, nil
	}

	v.lastNumber = latestNumber

	return true, nil
}

func getLatestAppNumber(log string) (int, error) {
	logEpoch, err := getLastAppLogLine(log)
	if err != nil {
		return -1, err
	}

	return logEpoch, nil
}

func getLastAppLogLine(log string) (int, error) {
	lines := strings.Split(log, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "[APP") {
			epoch, err := getLogEpoch(strings.TrimSpace(lines[i]))
			if err != nil {
				continue
			}

			return epoch, nil
		}
	}

	return -1, fmt.Errorf("cannot find any app logs")
}

func getLogEpoch(line string) (int, error) {
	outSplit := strings.SplitAfter(line, "OUT")
	afterOut := strings.TrimSpace(outSplit[len(outSplit)-1])

	return strconv.Atoi(afterOut)
}
