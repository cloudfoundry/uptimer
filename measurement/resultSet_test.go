package measurement_test

import (
	"time"

	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResultSet", func() {
	var (
		rs ResultSet
	)

	BeforeEach(func() {
		rs = NewResultSet()
	})

	It("generally works", func() {
		rs.RecordSuccess()
		rs.RecordFailure()

		Expect(rs.Successful()).To(Equal(1))
		Expect(rs.Failed()).To(Equal(1))
		Expect(rs.Total()).To(Equal(2))
	})

	It("records the time of the last failure", func() {
		rs.RecordFailure()
		time.Sleep(10 * time.Millisecond)

		now := time.Now().UTC()
		rs.RecordFailure()
		time.Sleep(10 * time.Millisecond)

		rs.RecordSuccess()
		rs.RecordSuccess()

		Expect(rs.LastFailure()).To(BeTemporally("~", now))
	})
})
