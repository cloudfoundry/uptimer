package measurement_test

import (
	"fmt"
	"net"

	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCPAvailability", func() {
	var (
		url  string
		port int

		am BaseMeasurement
	)

	BeforeEach(func() {
		url = "localhost"
		port = 6000 + GinkgoParallelProcess()

		am = NewTCPAvailability(url, port)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(am.Name()).To(Equal("TCP availability"))
		})
	})

	Describe("Summary Phrase", func() {
		It("returns the summary", func() {
			Expect(am.SummaryPhrase()).To(Equal("perform netcat requests"))
		})
	})

	Describe("PerformMeasurement", func() {
		var (
			listener net.Listener

			done chan any
		)

		Context("When the measurement client gets the expected response", func() {
			BeforeEach(func() {
				var err error
				listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
				Expect(err).NotTo(HaveOccurred())

				done = make(chan any)

				// Listen for an incoming connection.
				go func() {
					defer GinkgoRecover()

					conn, err := listener.Accept()
					Expect(err).NotTo(HaveOccurred())

					_, err = conn.Write([]byte("Hello from Uptimer."))
					Expect(err).NotTo(HaveOccurred())

					err = conn.Close()
					Expect(err).NotTo(HaveOccurred())

					close(done)
				}()
			})

			AfterEach(func() {
				err := listener.Close()
				Expect(err).NotTo(HaveOccurred())

				Eventually(done).Should(BeClosed())
			})

			It("records a matching string as success", func() {
				err, _, _, res := am.PerformMeasurement()

				Expect(err).To(Equal(""))
				Expect(res).To(BeTrue())
			})
		})

		Context("When the measurement client does not get the expected response", func() {
			BeforeEach(func() {
				var err error
				listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
				Expect(err).NotTo(HaveOccurred())

				done = make(chan any)

				// Listen for an incoming connection.
				go func() {
					defer GinkgoRecover()

					conn, err := listener.Accept()
					Expect(err).NotTo(HaveOccurred())

					_, err = conn.Write([]byte("Hello from Zuptimer."))
					Expect(err).NotTo(HaveOccurred())

					err = conn.Close()
					Expect(err).NotTo(HaveOccurred())

					close(done)
				}()
			})

			AfterEach(func() {
				err := listener.Close()
				Expect(err).NotTo(HaveOccurred())

				Eventually(done).Should(BeClosed())
			})

			It("records a mismatched string as failure", func() {
				err, _, _, res := am.PerformMeasurement()
				Expect(err).To(Equal("TCP App not returning expected response"))
				Expect(res).To(BeFalse())
			})
		})
	})
})
