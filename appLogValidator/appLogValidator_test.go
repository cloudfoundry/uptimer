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
			result, err := alv.IsNewer("[APP OUT 1500006820\n")

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})

		Context("the second time around...", func() {
			BeforeEach(func() {
				alv.IsNewer("[APP OUT 1500006820\n")
			})

			It("returns true if the next log's last line has a larger epoch value", func() {
				result, err := alv.IsNewer("[APP OUT 1500006821\n")

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})

			It("returns false if the next log's last line doesn't have a larger epoch value", func() {
				result, err := alv.IsNewer("[APP OUT 1500006819\n")

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})

			It("returns an error if called with a non-matching log line", func() {
				_, err := alv.IsNewer("")

				Expect(err).To(MatchError("cannot find any app logs"))
			})

			It("succeeds when there was a newer line even if the app exits", func() {
				result, err := alv.IsNewer("[APP OUT 1500006821\n[APP OUT Exit status 143\n")

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})

			It("returns false when there was not a newer line even if the app exits", func() {
				result, err := alv.IsNewer("[APP OUT 1500006819\n[APP OUT Exit status 143\n")

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})

			It("returns an error if called with a log line that doesn't have an epoch", func() {
				_, err := alv.IsNewer("[APP OUT notAnEpoch\n")

				Expect(err.Error()).To(Equal(`cannot find any app logs`))
			})
		})

		Context("when the last line is not a number", func() {
			BeforeEach(func() {
				alv.IsNewer("[APP OUT 1510680926\n")
			})

			It("returns the last number line", func() {
				result, err := alv.IsNewer(`2017-11-14T09:35:27.02-0800 [APP/PROC/WEB/0] OUT Starting health monitoring of container
2017-11-14T09:35:27.02-0800 [APP/PROC/WEB/0] OUT Uploading complete
2017-11-14T09:35:27.02-0800 [APP/PROC/WEB/0] OUT Stopping instance 260a079c-71a7-4c8e-970e-981881836c23
2017-11-14T09:35:27.02-0800 [APP/PROC/WEB/0] OUT Destroying container
2017-11-14T09:35:27.02-0800 [APP/PROC/WEB/0] OUT Creating container
2017-11-14T09:35:28.02-0800 [APP/PROC/WEB/0] OUT 1510680925
2017-11-14T09:35:28.02-0800 [APP/PROC/WEB/1] OUT Starting health monitoring of container
2017-11-14T09:35:29.02-0800 [APP/PROC/WEB/0] OUT 1510680926
2017-11-14T09:35:29.02-0800 [APP/PROC/WEB/0] OUT 1510680927
2017-11-14T09:35:29.04-0800 [APP/PROC/WEB/1] OUT Container became healthy`)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})
	})
})
