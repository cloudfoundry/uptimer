package measurement_test

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	. "github.com/cloudfoundry/uptimer/measurement"

	"bytes"
	"io/ioutil"

	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCPAvailability", func() {
	var (
		url              string
		port             int
		fakeRoundTripper *FakeRoundTripper
		successResponse  *http.Response
		failResponse     *http.Response
		client           *http.Client

		am BaseMeasurement
	)

	BeforeEach(func() {
		url = "tcp.example.com"
		port = 6000
		fakeRoundTripper = &FakeRoundTripper{}
		successResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}
		failResponse = &http.Response{
			StatusCode: 400,
			Status:     "Bad Request",
			Body:       ioutil.NopCloser(bytes.NewBufferString("Body of the error here")),
		}
		fakeRoundTripper.RoundTripReturns(successResponse, nil)
		client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: fakeRoundTripper,
		}

		am = NewHTTPAvailability(url, client)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(am.Name()).To(Equal("HTTP availability"))
		})
	})

	Describe("PerformMeasurement", func() {
		It("makes a nc request to the url", func() {
			am.PerformMeasurement()

			req := fakeRoundTripper.RoundTripArgsForCall(0)
			Expect(req.Method).To(Equal(http.MethodGet))
			Expect(req.URL.String()).To(Equal("https://example.com/foo"))
		})

		It("records a matching string as success", func() {
			_, _, _, res := am.PerformMeasurement()

			Expect(res).To(BeTrue())
		})

		It("records an error response results as failed", func() {
			fakeRoundTripper.RoundTripReturns(failResponse, nil)

			_, _, _, res := am.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("records a mismatched string as failed", func() {
			fakeRoundTripper.RoundTripReturns(nil, fmt.Errorf("error"))

			_, _, _, res := am.PerformMeasurement()

			Expect(res).To(BeFalse())
		})


		It("closes the body of the response when there is a 200 response", func() {
			fakeRC := &fakeReadCloser{}
			fakeRoundTripper.RoundTripReturns(
				&http.Response{
					StatusCode: 200,
					Body:       fakeRC,
				},
				nil,
			)

			am.PerformMeasurement()

			Expect(fakeRC.Closed).To(BeTrue())
		})

		It("closes the body of the response when there is a non-200 response", func() {
			fakeRC := &fakeReadCloser{}
			fakeRoundTripper.RoundTripReturns(
				&http.Response{
					StatusCode: 400,
					Body:       fakeRC,
				},
				nil,
			)

			am.PerformMeasurement()

			Expect(fakeRC.Closed).To(BeTrue())
		})

		It("does not close the body of the response when there is an error", func() {
			fakeRoundTripper.RoundTripReturns(
				nil,
				fmt.Errorf("foobar"),
			)

			Expect(func() { am.PerformMeasurement() }).NotTo(Panic())
		})
	})
})

type fakeReadCloser struct {
	Closed bool
}

func (rc *fakeReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (rc *fakeReadCloser) Close() error {
	rc.Closed = true

	return nil
}

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
