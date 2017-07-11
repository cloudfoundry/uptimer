package measurement_test

import (
	"fmt"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/fakes"
	. "github.com/cloudfoundry/uptimer/measurement"

	"time"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Availability", func() {
	var (
		url              string
		freq             time.Duration
		mockClock        *clock.Mock
		fakeRoundTripper *fakes.FakeRoundTripper
		successResponse  *http.Response
		failResponse     *http.Response
		client           *http.Client

		am Measurement
	)

	BeforeEach(func() {
		url = "https://example.com/foo"
		freq = time.Second
		mockClock = clock.NewMock()
		fakeRoundTripper = &fakes.FakeRoundTripper{}
		successResponse = &http.Response{StatusCode: 200}
		failResponse = &http.Response{StatusCode: 400}
		fakeRoundTripper.RoundTripReturns(successResponse, nil)
		client = &http.Client{
			Transport: fakeRoundTripper,
		}

		am = NewAvailability(url, freq, mockClock, client)
	})

	Describe("Start", func() {
		AfterEach(func() {
			am.Stop()
		})

		It("makes a get request to the url", func() {
			err := am.Start()
			mockClock.Add(freq)
			Expect(err).NotTo(HaveOccurred())

			req := fakeRoundTripper.RoundTripArgsForCall(0)
			Expect(req.Method).To(Equal(http.MethodGet))
			Expect(req.URL.String()).To(Equal("https://example.com/foo"))
		})

		It("makes a request with given frequency", func() {

			am.Start()
			mockClock.Add(3 * freq)

			Expect(fakeRoundTripper.RoundTripCallCount()).To(Equal(4))
		})

		It("records the total results", func() {
			am.Start()
			mockClock.Add(3 * freq)

			rs, _ := am.Results()
			Expect(rs.Total()).To(Equal(4))
			Expect(rs.Successful()).To(Equal(4))
			Expect(rs.Failed()).To(Equal(0))
		})

		It("records the non-200 results as failed", func() {
			fakeRoundTripper.RoundTripReturns(failResponse, nil)

			am.Start()
			mockClock.Add(3 * freq)

			rs, _ := am.Results()
			Expect(rs.Total()).To(Equal(4))
			Expect(rs.Successful()).To(Equal(0))
			Expect(rs.Failed()).To(Equal(4))
		})

		It("records the error results as failed", func() {
			fakeRoundTripper.RoundTripReturns(nil, fmt.Errorf("error"))

			am.Start()
			mockClock.Add(3 * freq)

			rs, _ := am.Results()
			Expect(rs.Total()).To(Equal(4))
			Expect(rs.Successful()).To(Equal(0))
			Expect(rs.Failed()).To(Equal(4))
		})
	})

	Describe("Stop", func() {
		It("stops the measurement", func() {
			am.Start()
			mockClock.Add(3 * freq)
			am.Stop()
			mockClock.Add(3 * freq)

			Expect(fakeRoundTripper.RoundTripCallCount()).To(Equal(4))
		})
	})
})
