package measurement_test

import (
	"fmt"
	"net"
	"os"

	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCPAvailability", func() {
	var (
		url  string
		port int

		am BaseMeasurement

		server net.Listener
		conn   net.Conn
		err    error
	)

	BeforeEach(func() {
		url = "localhost"
		port = 6000
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(am.Name()).To(Equal("TCP availability"))
		})
	})

	Describe("PerformMeasurement", func() {
		It("makes a nc request to the url", func() {
			// am.PerformMeasurement()

			// req := fakeRoundTripper.RoundTripArgsForCall(0)
			// Expect(req.Method).To(Equal(http.MethodGet))
			// Expect(req.URL.String()).To(Equal("https://example.com/foo"))
		})

		Context("When the measurement client gets the expected response", func() {
			BeforeEach(func() {
				go func() {
					server, err = net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
					if err != nil {
						fmt.Println("Error listening: ", err.Error())
						os.Exit(1)
					}
					conn, err = server.Accept()
					if err != nil {
						fmt.Println("Error accepting: ", err.Error())
						os.Exit(1)
					}

					_, err = conn.Write([]byte("Hello from Uptimer."))
					if err != nil {
						fmt.Println("Error writing:", err.Error())
						os.Exit(1)
					}
					Expect(err).ToNot(HaveOccurred())
				}()

				am = NewTCPAvailability(url, port)
			})

			AfterEach(func() {
				server.Close()
				conn.Close()
			})
			FIt("records a matching string as success", func() {

				err, _, _, res := am.PerformMeasurement()

				Expect(err).To(Equal(""))
				Expect(res).To(BeTrue())
			})
		})

		It("records an error response results as failed", func() {
			// fakeRoundTripper.RoundTripReturns(failResponse, nil)

			// _, _, _, res := am.PerformMeasurement()

			// Expect(res).To(BeFalse())
		})

		It("records a mismatched string as failed", func() {
			// fakeRoundTripper.RoundTripReturns(nil, fmt.Errorf("error"))

			// _, _, _, res := am.PerformMeasurement()

			// Expect(res).To(BeFalse())
		})

		It("closes the body of the response when there is a 200 response", func() {
			// fakeRC := &fakeReadCloser{}
			// fakeRoundTripper.RoundTripReturns(
			// 	&http.Response{
			// 		StatusCode: 200,
			// 		Body:       fakeRC,
			// 	},
			// 	nil,
			// )

			// am.PerformMeasurement()

			// Expect(fakeRC.Closed).To(BeTrue())
		})

		It("closes the body of the response when there is a non-200 response", func() {
			// fakeRC := &fakeReadCloser{}
			// fakeRoundTripper.RoundTripReturns(
			// 	&http.Response{
			// 		StatusCode: 400,
			// 		Body:       fakeRC,
			// 	},
			// 	nil,
			// )

			// am.PerformMeasurement()

			// Expect(fakeRC.Closed).To(BeTrue())
		})

		It("does not close the body of the response when there is an error", func() {
			// fakeRoundTripper.RoundTripReturns(
			// 	nil,
			// 	fmt.Errorf("foobar"),
			// )

			// Expect(func() { am.PerformMeasurement() }).NotTo(Panic())
		})
	})
})
