package appLogValidator_test

import (
	. "github.com/cloudfoundry/uptimer/appLogValidator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppLogValidator", func() {
	Describe("IsNewer", func() {
		var (
			alv AppLogValidator
		)

		BeforeEach(func() {
			alv = New()
		})

		It("always returns true the first time it's called", func() {
			result, err := alv.IsNewer("[APP OUT 1500006820")

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})

		Context("the second time around...", func() {
			BeforeEach(func() {
				alv.IsNewer("[APP OUT 1500006820")
			})

			It("returns true if the next log's last line has a larger epoch value", func() {
				result, err := alv.IsNewer("[APP OUT 1500006821")

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})

			It("returns false if the next log's last line doesn't have a larger epoch value", func() {
				result, err := alv.IsNewer("[APP OUT 1500006819")

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})

			It("returns an error if called with a non-matching log line", func() {
				_, err := alv.IsNewer("")

				Expect(err).To(MatchError("cannot find any app logs"))
			})

			It("returns a failure when the app exits", func() {
				_, err := alv.IsNewer("[APP OUT Exit status 143")

				Expect(err).To(MatchError("app exited"))
			})

			It("returns an error if called with a log line that doesn't have an epoch", func() {
				_, err := alv.IsNewer("[APP OUT notAnEpoch")

				Expect(err.Error()).To(Equal(`strconv.Atoi: parsing "notAnEpoch": invalid syntax`))
			})
		})
	})
})
