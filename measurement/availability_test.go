package measurement_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Availability", func() {
	var (
		url              string
		freq             time.Duration
		mockClock        *clock.Mock
		fakeRoundTripper *FakeRoundTripper
		successResponse  *http.Response
		failResponse     *http.Response
		client           *http.Client
		logger           *log.Logger
		logBuf           *bytes.Buffer

		am Measurement
	)

	BeforeEach(func() {
		url = "https://example.com/foo"
		freq = time.Second
		mockClock = clock.NewMock()
		fakeRoundTripper = &FakeRoundTripper{}
		successResponse = &http.Response{StatusCode: 200}
		failResponse = &http.Response{StatusCode: 400}
		fakeRoundTripper.RoundTripReturns(successResponse, nil)
		client = &http.Client{
			Transport: fakeRoundTripper,
		}
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)

		am = NewAvailability(logger, url, freq, mockClock, client)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(am.Name()).To(Equal("HTTP availability"))
		})
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

		It("logs error output when there is a non-200 response", func() {
			fakeRoundTripper.RoundTripReturns(failResponse, nil)

			am.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(HTTP availability): response had status 400\x1b[0m\n"))
		})

		It("logs error output when there is an error", func() {
			fakeRoundTripper.RoundTripReturns(nil, fmt.Errorf("error"))

			am.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(HTTP availability): Get https://example.com/foo: error\x1b[0m\n"))
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

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			am.Start()
			mockClock.Add(3 * freq)

			Expect(am.Failed()).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			am.Start()
			fakeRoundTripper.RoundTripReturns(failResponse, nil)
			mockClock.Add(3 * freq)

			Expect(am.Failed()).To(BeTrue())
		})
	})

	Describe("Summary", func() {
		It("returns a success summary if none failed", func() {
			am.Start()
			mockClock.Add(3 * freq)
			am.Stop()

			Expect(am.Summary()).To(Equal(fmt.Sprintf("SUCCESS(%s): All %d requests succeeded", am.Name(), 4)))
		})
		It("returns a failed summary if there are failures", func() {
			am.Start()
			mockClock.Add(3 * freq)
			fakeRoundTripper.RoundTripReturns(failResponse, nil)
			mockClock.Add(3 * freq)
			am.Stop()

			Expect(am.Summary()).To(Equal(fmt.Sprintf("FAILED(%s): %d of %d requests failed", am.Name(), 3, 7)))
		})
	})
})

type FakeRoundTripper struct {
	RoundTripStub        func(*http.Request) (*http.Response, error)
	roundTripMutex       sync.RWMutex
	roundTripArgsForCall []struct {
		arg1 *http.Request
	}
	roundTripReturns struct {
		result1 *http.Response
		result2 error
	}
}

func (fake *FakeRoundTripper) RoundTrip(arg1 *http.Request) (*http.Response, error) {
	fake.roundTripMutex.Lock()
	fake.roundTripArgsForCall = append(fake.roundTripArgsForCall, struct {
		arg1 *http.Request
	}{arg1})
	fake.roundTripMutex.Unlock()
	if fake.RoundTripStub != nil {
		return fake.RoundTripStub(arg1)
	} else {
		return fake.roundTripReturns.result1, fake.roundTripReturns.result2
	}
}

func (fake *FakeRoundTripper) RoundTripCallCount() int {
	fake.roundTripMutex.RLock()
	defer fake.roundTripMutex.RUnlock()
	return len(fake.roundTripArgsForCall)
}

func (fake *FakeRoundTripper) RoundTripArgsForCall(i int) *http.Request {
	fake.roundTripMutex.RLock()
	defer fake.roundTripMutex.RUnlock()
	return fake.roundTripArgsForCall[i].arg1
}

func (fake *FakeRoundTripper) RoundTripReturns(result1 *http.Response, result2 error) {
	fake.RoundTripStub = nil
	fake.roundTripReturns = struct {
		result1 *http.Response
		result2 error
	}{result1, result2}
}
