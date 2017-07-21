package measurement

import (
	"sort"
	"time"
)

//go:generate counterfeiter . ResultSet
type ResultSet interface {
	RecordSuccess()
	RecordFailure()

	SuccessesSinceLastFailure() (int, time.Time)

	Successful() int
	Failed() int
	Total() int
}

type resultSet struct {
	successful []time.Time
	failed     []time.Time
}

func NewResultSet() ResultSet {
	return &resultSet{}
}

func (rs *resultSet) RecordSuccess() {
	rs.successful = append(rs.successful, time.Now().UTC())
}

func (rs *resultSet) RecordFailure() {
	rs.failed = append(rs.failed, time.Now().UTC())
}

func (rs *resultSet) Successful() int {
	return len(rs.successful)
}

func (rs *resultSet) Failed() int {
	return len(rs.failed)
}

func (rs *resultSet) Total() int {
	return len(rs.successful) + len(rs.failed)
}

func (rs *resultSet) SuccessesSinceLastFailure() (int, time.Time) {
	if len(rs.successful) == 0 || len(rs.failed) == 0 {
		return 0, time.Time{}
	}

	lf := rs.failed[len(rs.failed)-1]
	s := sort.Search(len(rs.successful), func(i int) bool { return rs.successful[i].After(lf) })

	if s < len(rs.failed) {
		return len(rs.successful) - s, lf
	}

	return 0, time.Time{}
}
