package timeshim

import "time"

//go:generate counterfeiter -o time_fake/fake_time.go . Time
type Time interface {
	Now() time.Time
}
