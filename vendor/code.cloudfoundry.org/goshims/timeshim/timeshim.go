package timeshim

import "time"

type TimeShim struct{}

func (sh *TimeShim) Now() time.Time {
	return time.Now()
}
