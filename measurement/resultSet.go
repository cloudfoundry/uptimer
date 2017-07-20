package measurement

//go:generate counterfeiter . ResultSet
type ResultSet interface {
	RecordSuccess()
	RecordFailure()

	Successful() int
	Failed() int
	Total() int
}

type resultSet struct {
	successful int
	failed     int
}

func NewResultSet() ResultSet {
	return &resultSet{}
}

func (rs *resultSet) RecordSuccess() {
	rs.successful++
}

func (rs *resultSet) RecordFailure() {
	rs.failed++
}

func (rs *resultSet) Successful() int {
	return rs.successful
}

func (rs *resultSet) Failed() int {
	return rs.failed
}

func (rs *resultSet) Total() int {
	return rs.successful + rs.failed
}
