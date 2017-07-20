package measurement_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"

	. "github.com/cloudfoundry/uptimer/measurement"
	"github.com/cloudfoundry/uptimer/measurement/measurementfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Availability", func() {
	var (
		url              string
		fakeRoundTripper *FakeRoundTripper
		successResponse  *http.Response
		failResponse     *http.Response
		client           *http.Client
		logger           *log.Logger
		logBuf           *bytes.Buffer
		fakeResultSet    *measurementfakes.FakeResultSet

		am BaseMeasurement
	)

	BeforeEach(func() {
		url = "https://example.com/foo"
		fakeRoundTripper = &FakeRoundTripper{}
		successResponse = &http.Response{StatusCode: 200}
		failResponse = &http.Response{StatusCode: 400}
		fakeRoundTripper.RoundTripReturns(successResponse, nil)
		client = &http.Client{
			Transport: fakeRoundTripper,
		}
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)
		fakeResultSet = &measurementfakes.FakeResultSet{}

		am = NewAvailability(url, client)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(am.Name()).To(Equal("HTTP availability"))
		})
	})

	Describe("PerformMeasurement", func() {
		It("makes a get request to the url", func() {
			am.PerformMeasurement(logger)

			req := fakeRoundTripper.RoundTripArgsForCall(0)
			Expect(req.Method).To(Equal(http.MethodGet))
			Expect(req.URL.String()).To(Equal("https://example.com/foo"))
		})

		It("records 200 results as success", func() {
			res := am.PerformMeasurement(logger)

			Expect(res).To(BeTrue())
		})

		It("records the non-200 results as failed", func() {
			fakeRoundTripper.RoundTripReturns(failResponse, nil)

			res := am.PerformMeasurement(logger)

			Expect(res).To(BeFalse())
		})

		It("records the error results as failed", func() {
			fakeRoundTripper.RoundTripReturns(nil, fmt.Errorf("error"))

			res := am.PerformMeasurement(logger)

			Expect(res).To(BeFalse())
		})

		It("logs error output when there is a non-200 response", func() {
			fakeRoundTripper.RoundTripReturns(failResponse, nil)

			am.PerformMeasurement(logger)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(HTTP availability): response had status 400\x1b[0m\n"))
		})

		It("logs error output when there is an error", func() {
			fakeRoundTripper.RoundTripReturns(nil, fmt.Errorf("error"))

			am.PerformMeasurement(logger)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(HTTP availability): Get https://example.com/foo: error\x1b[0m\n"))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			fakeResultSet.FailedReturns(0)

			am.PerformMeasurement(logger)

			Expect(am.Failed(fakeResultSet)).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			fakeResultSet.FailedReturns(1)

			am.PerformMeasurement(logger)

			Expect(am.Failed(fakeResultSet)).To(BeTrue())
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
