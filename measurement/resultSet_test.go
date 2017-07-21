package measurement_test

import (
	"time"

	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("ResultSet", func() {
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

	Describe("LastFailure", func() {
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

	Describe("SuccessesSinceLastFailure", func() {
		It("returns the number of successes since the last failure", func() {
			rs.RecordFailure()

			rs.RecordSuccess()
			rs.RecordSuccess()
			rs.RecordSuccess()

			Expect(rs.SuccessesSinceLastFailure()).To(Equal(3))
		})

		It("returns 0 if there have been no successes since the last failure", func() {
			rs.RecordFailure()

			rs.RecordSuccess()
			rs.RecordSuccess()
			rs.RecordSuccess()

			rs.RecordFailure()

			Expect(rs.SuccessesSinceLastFailure()).To(BeZero())
		})
	})
})
