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

		listener net.Listener
	)

	BeforeEach(func() {
		url = "localhost"
		port = 6000

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
		Context("When the measurement client gets the expected response", func() {
			BeforeEach(func() {
				var err error
				listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
				if err != nil {
					fmt.Println("Error listening: ", err.Error())
					os.Exit(1)
				}

				// Listen for an incoming connection.
				go func() {
					conn, err := listener.Accept()
					if err != nil {
						fmt.Println("Error accepting: ", err.Error())
						os.Exit(1)
					}
					// Handle connections in a new goroutine.
					go func(conn net.Conn) {
						defer conn.Close()

						_, err := conn.Write([]byte("Hello from Uptimer."))
						if err != nil {
							fmt.Println("Error writing:", err.Error())
							os.Exit(1)
						}
					}(conn)
				}()
			})

			AfterEach(func() {
				listener.Close()
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
				if err != nil {
					fmt.Println("Error listening: ", err.Error())
					os.Exit(1)
				}

				// Listen for an incoming connection.
				go func() {
					conn, err := listener.Accept()
					if err != nil {
						fmt.Println("Error accepting: ", err.Error())
						os.Exit(1)
					}
					// Handle connections in a new goroutine.
					go func(conn net.Conn) {
						defer conn.Close()

						_, err := conn.Write([]byte("Hello from Zuptimer."))
						if err != nil {
							fmt.Println("Error writing:", err.Error())
							os.Exit(1)
						}
					}(conn)
				}()
			})

			It("records a mismatched string as failure", func() {
				err, _, _, res := am.PerformMeasurement()

				listener.Close()
				Expect(err).To(Equal("TCP App not returning expected response"))
				Expect(res).To(BeFalse())
			})
		})
	})
})
