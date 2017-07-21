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

	Describe("SuccessesSinceLastFailure", func() {
		It("returns the number of successes since the last failure", func() {
			rs.RecordFailure()

			rs.RecordSuccess()
			rs.RecordSuccess()
			rs.RecordSuccess()

			s, _ := rs.SuccessesSinceLastFailure()

			Expect(s).To(Equal(3))
		})

		It("records the time of the last failure", func() {
			rs.RecordFailure()
			time.Sleep(10 * time.Millisecond)

			lastFailure := time.Now().UTC()
			rs.RecordFailure()
			time.Sleep(10 * time.Millisecond)

			rs.RecordSuccess()
			rs.RecordSuccess()

			s, t := rs.SuccessesSinceLastFailure()

			Expect(s).To(Equal(2))
			Expect(t).To(BeTemporally("~", lastFailure))
		})

		It("returns 0 and a base time value when there have been no successes", func() {
			rs.RecordFailure()
			rs.RecordFailure()
			rs.RecordFailure()

			s, t := rs.SuccessesSinceLastFailure()

			Expect(s).To(BeZero())
			Expect(t).To(Equal(time.Time{}))
		})

		It("returns 0 and a base time value when there have been no failures", func() {
			rs.RecordSuccess()
			rs.RecordSuccess()
			rs.RecordSuccess()

			s, t := rs.SuccessesSinceLastFailure()

			Expect(s).To(BeZero())
			Expect(t).To(Equal(time.Time{}))
		})

		It("returns 0 and a base time value if there have been no successes since the last failure", func() {
			rs.RecordFailure()

			rs.RecordSuccess()
			rs.RecordSuccess()
			rs.RecordSuccess()

			rs.RecordFailure()

			s, t := rs.SuccessesSinceLastFailure()

			Expect(s).To(BeZero())
			Expect(t).To(Equal(time.Time{}))
		})
	})
})
