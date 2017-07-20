package measurement_test

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/benbjohnson/clock"
	. "github.com/cloudfoundry/uptimer/measurement"

	"github.com/cloudfoundry/uptimer/measurement/measurementfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Periodic", func() {
	var (
		logger              *log.Logger
		mockClock           *clock.Mock
		freq                time.Duration
		fakeBaseMeasurement *measurementfakes.FakeBaseMeasurement
		fakeResultSet       *measurementfakes.FakeResultSet

		p Measurement
	)

	BeforeEach(func() {
		logger = log.New(ioutil.Discard, "", 0)
		mockClock = clock.NewMock()
		freq = time.Second
		fakeBaseMeasurement = &measurementfakes.FakeBaseMeasurement{}
		fakeResultSet = &measurementfakes.FakeResultSet{}

		p = NewPeriodic(logger, mockClock, freq, fakeBaseMeasurement, fakeResultSet)
	})

	Describe("Name", func() {
		It("Returns the base measurement's name", func() {
			fakeBaseMeasurement.NameReturns("foo measurement")

			Expect(p.Name()).To(Equal("foo measurement"))
		})
	})

	Describe("Start", func() {
		AfterEach(func() {
			p.Stop()
		})

		It("runs the base measurement immediately, before one frequency elapses", func() {
			p.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(1))
		})

		It("runs the base measurement with given frequency", func() {
			p.Start()
			mockClock.Add(3 * freq)

			Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(4))
		})
	})

	Describe("Stop", func() {
		It("stops the measurement", func() {
			p.Start()
			mockClock.Add(3 * freq)
			p.Stop()
			mockClock.Add(3 * freq)

			Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(4))
		})
	})

	Describe("Results", func() {
		It("returns the result set", func() {
			Expect(p.Results()).To(Equal(fakeResultSet))
		})
	})

	Describe("Failed", func() {
		It("Returns the base measurement's failed state", func() {
			fakeBaseMeasurement.FailedReturns(true)

			Expect(p.Failed()).To(BeTrue())
		})
	})

	Describe("Summary", func() {
		It("Returns the base measurement's summary", func() {
			fakeBaseMeasurement.SummaryReturns("this is a summary")

			Expect(p.Summary()).To(Equal("this is a summary"))
		})
	})
})
