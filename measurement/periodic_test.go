package measurement_test

import (
	"bytes"
	"fmt"
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
		logBuf              *bytes.Buffer
		logger              *log.Logger
		mockClock           *clock.Mock
		freq                time.Duration
		fakeBaseMeasurement *measurementfakes.FakeBaseMeasurement
		fakeResultSet       *measurementfakes.FakeResultSet

		p Measurement
	)

	BeforeEach(func() {
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)
		mockClock = clock.NewMock()
		freq = time.Second
		fakeBaseMeasurement = &measurementfakes.FakeBaseMeasurement{}
		fakeBaseMeasurement.NameReturns("foo measurement")
		fakeBaseMeasurement.SummaryPhraseReturns("wingdang the foobrizzle")
		fakeResultSet = &measurementfakes.FakeResultSet{}

		p = NewPeriodic(logger, mockClock, freq, fakeBaseMeasurement, fakeResultSet)
	})

	Describe("Name", func() {
		It("Returns the base measurement's name", func() {
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

		It("records success if the base measurement succeeds", func() {
			fakeBaseMeasurement.PerformMeasurementReturns("", true)

			p.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(1))
			Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(0))
		})

		It("records failure if the base measurement fails", func() {
			fakeBaseMeasurement.PerformMeasurementReturns("", false)

			p.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(0))
			Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(1))
		})

		It("logs when the measurement fails", func() {
			fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", false)

			p.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("measurement failed!\n"))
		})

		It("logs how many successes since the last failure", func() {
			lastFailure := mockClock.Now().UTC()
			fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", false)
			fakeResultSet.SuccessesSinceLastFailureReturns(3, lastFailure)

			p.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal(fmt.Sprintf("measurement failed!\n3 successes since last failure (at %s)\n", lastFailure.Format("2006/01/02 15:04:05"))))
		})

		It("does not logs how many successes since the last failure if there have been none", func() {
			fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", false)
			fakeResultSet.SuccessesSinceLastFailureReturns(0, time.Time{})

			p.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("measurement failed!\n"))
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
		It("returns a success summary if none failed", func() {
			fakeBaseMeasurement.FailedReturns(false)
			fakeResultSet.FailedReturns(0)
			fakeResultSet.TotalReturns(4)

			Expect(p.Summary()).To(Equal(fmt.Sprintf("SUCCESS(%s): All %d attempts to %s succeeded", fakeBaseMeasurement.Name(), 4, fakeBaseMeasurement.SummaryPhrase())))
		})

		It("returns a failed summary if there are failures", func() {
			fakeBaseMeasurement.FailedReturns(true)
			fakeResultSet.FailedReturns(3)
			fakeResultSet.TotalReturns(7)

			Expect(p.Summary()).To(Equal(fmt.Sprintf("FAILED(%s): %d of %d attempts to %s failed", fakeBaseMeasurement.Name(), 3, 7, fakeBaseMeasurement.SummaryPhrase())))
		})

		It("returns a failed summary with additional last x succeeded if there are successes since the last failure", func() {
			fakeBaseMeasurement.FailedReturns(true)
			fakeResultSet.FailedReturns(3)
			fakeResultSet.TotalReturns(7)
			fakeResultSet.SuccessesSinceLastFailureReturns(2, time.Time{})

			Expect(p.Summary()).To(Equal(
				fmt.Sprintf(
					"FAILED(%s): %d of %d attempts to %s failed (the last %d succeeded)",
					fakeBaseMeasurement.Name(),
					3,
					7,
					fakeBaseMeasurement.SummaryPhrase(),
					2,
				)))
		})
	})
})
