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
		allowedFailures     int
		shouldRetry         bool

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
		allowedFailures = 0
		shouldRetry = false

		p = NewPeriodic(
			logger,
			mockClock,
			freq,
			fakeBaseMeasurement,
			fakeResultSet,
			allowedFailures,
			func(string, string) bool { return shouldRetry },
		)
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
		Context("with measure immediately", func() {
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

		Context("without measure immediately", func() {
			BeforeEach(func() {
				p = NewPeriodicWithoutMeasuringImmediately(
					logger,
					mockClock,
					freq,
					fakeBaseMeasurement,
					fakeResultSet,
					allowedFailures,
					func(string, string) bool { return shouldRetry },
				)
			})

			It("does not measure until one time frequency has passed", func() {
				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(0))
			})

			It("runs the base measurement with given frequency, starting after one frequency has passed", func() {
				p.Start()
				mockClock.Add(3 * freq)

				Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(3))
			})
		})

		Context("when shouldRetry is false", func() {
			BeforeEach(func() {
				fakeResultSet.FailedReturns(2)
				allowedFailures = 4

				p = NewPeriodic(
					logger,
					mockClock,
					freq,
					fakeBaseMeasurement,
					fakeResultSet,
					allowedFailures,
					func(string, string) bool { return shouldRetry },
				)
			})

			It("records success if the base measurement succeeds", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("", "", "", true)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(1))
				Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(0))
			})

			It("records failure if the base measurement fails", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("", "", "", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(0))
				Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(1))
			})

			It("logs when the measurement fails without stdout or stderr", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", "", "", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE (foo measurement, 2/4): measurement failed!\x1b[0m\n\n"))
			})

			It("logs when the measurement fails with stdout", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", "out out!", "", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE (foo measurement, 2/4): measurement failed!\x1b[0m\n\nstdout:\nout out!\n\n"))
			})

			It("logs when the measurement fails with stderr", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", "", "err err!", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE (foo measurement, 2/4): measurement failed!\x1b[0m\n\nstderr:\nerr err!\n\n"))
			})

			It("logs when the measurement fails with both stdout or stderr", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", "out out!", "err err!", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE (foo measurement, 2/4): measurement failed!\x1b[0m\n\nstdout:\nout out!\n\nstderr:\nerr err!\n\n"))
			})

			It("logs how many successes since the last failure", func() {
				lastFailure := mockClock.Now().UTC()
				fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", "", "", false)
				fakeResultSet.SuccessesSinceLastFailureReturns(3, lastFailure)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal(fmt.Sprintf("\x1b[31mFAILURE (foo measurement, 2/4): measurement failed! (3 successes since last failure at %s)\x1b[0m\n\n", lastFailure.Format("2006/01/02 15:04:05"))))
			})

			It("does not logs how many successes since the last failure if there have been none", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("measurement failed!", "", "", false)
				fakeResultSet.SuccessesSinceLastFailureReturns(0, time.Time{})

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE (foo measurement, 2/4): measurement failed!\x1b[0m\n\n"))
			})
		})

		Context("when should retry is true", func() {
			BeforeEach(func() {
				shouldRetry = true
				fakeResultSet.FailedReturns(1)
				allowedFailures = 2

				p = NewPeriodic(
					logger,
					mockClock,
					freq,
					fakeBaseMeasurement,
					fakeResultSet,
					allowedFailures,
					func(string, string) bool { return shouldRetry },
				)
			})

			It("runs the base measurement only once if perform measurement returns ok the first time", func() {
				fakeBaseMeasurement.PerformMeasurementReturns("", "", "", true)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(1))
			})

			It("runs the base measurement twice if perform measurement returns not ok and then ok", func() {
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "", "", "", true)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeBaseMeasurement.PerformMeasurementCallCount()).To(Equal(2))
			})

			It("records success if perform measurement returns not ok and then ok", func() {
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "", "", "", true)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(1))
				Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(0))
			})

			It("does not log if perform measurement returns not ok and then ok", func() {
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "", "", "", true)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(BeEmpty())
			})

			It("records failure if perform measurement returns not ok twice", func() {
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "", "", "", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(0))
				Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(1))
			})

			It("logs failure if perform measurement returns not ok twice", func() {
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "", "", "", false)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(ContainSubstring("FAILURE (foo measurement, 1/2)"))
			})

			It("logs how many successes since the last failure", func() {
				lastFailure := mockClock.Now().UTC()
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "measurement failed!", "", "", false)
				fakeResultSet.SuccessesSinceLastFailureReturns(3, lastFailure)

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal(fmt.Sprintf("\x1b[31mFAILURE (foo measurement, 1/2): measurement failed! (3 successes since last failure at %s)\x1b[0m\n\n", lastFailure.Format("2006/01/02 15:04:05"))))
			})

			It("does not logs how many successes since the last failure if there have been none", func() {
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(0, "", "", "", false)
				fakeBaseMeasurement.PerformMeasurementReturnsOnCall(1, "measurement failed!", "", "", false)
				fakeResultSet.SuccessesSinceLastFailureReturns(0, time.Time{})

				p.Start()
				mockClock.Add(freq - time.Nanosecond)

				Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE (foo measurement, 1/2): measurement failed!\x1b[0m\n\n"))
			})
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
		BeforeEach(func() {
			p = NewPeriodic(logger, mockClock, freq, fakeBaseMeasurement, fakeResultSet, 5, func(string, string) bool { return shouldRetry })
		})

		It("Returns true if failure count > allowed number of failures", func() {
			fakeResultSet.FailedReturns(6)

			Expect(p.Failed()).To(BeTrue())
		})

		It("Returns false if failure count < allowed number of failures", func() {
			fakeResultSet.FailedReturns(4)

			Expect(p.Failed()).To(BeFalse())
		})

		It("Returns false if failure count == allowed number of failures", func() {
			fakeResultSet.FailedReturns(5)

			Expect(p.Failed()).To(BeFalse())
		})
	})

	Describe("Summary", func() {
		It("returns a success summary if fewer than the allowed failures occurred", func() {
			failed := 2
			succeeded := 2
			allowedFailures := 3

			p = NewPeriodic(logger, mockClock, freq, fakeBaseMeasurement, fakeResultSet, allowedFailures, func(string, string) bool { return shouldRetry })
			fakeResultSet.FailedReturns(failed)
			fakeResultSet.TotalReturns(failed + succeeded)

			Expect(p.Summary()).To(Equal(
				fmt.Sprintf("SUCCESS (%s): %d failed attempts to %s did not exceed the threshold of %d allowed failures (Total attempts: %d)",
					fakeBaseMeasurement.Name(),
					failed,
					fakeBaseMeasurement.SummaryPhrase(),
					allowedFailures,
					failed+succeeded,
				)))
		})

		It("returns a success summary if exactly the number of allowed failures occurred", func() {
			failed := 2
			succeeded := 2
			allowedFailures := 2

			p = NewPeriodic(logger, mockClock, freq, fakeBaseMeasurement, fakeResultSet, allowedFailures, func(string, string) bool { return shouldRetry })
			fakeResultSet.FailedReturns(failed)
			fakeResultSet.TotalReturns(failed + succeeded)

			Expect(p.Summary()).To(Equal(
				fmt.Sprintf("SUCCESS (%s): %d failed attempts to %s did not exceed the threshold of %d allowed failures (Total attempts: %d)",
					fakeBaseMeasurement.Name(),
					failed,
					fakeBaseMeasurement.SummaryPhrase(),
					allowedFailures,
					failed+succeeded,
				)))
		})

		It("returns a failed summary if more than the allowed failures occurred", func() {
			failed := 3
			succeeded := 2
			allowedFailures := 2

			p = NewPeriodic(logger, mockClock, freq, fakeBaseMeasurement, fakeResultSet, allowedFailures, func(string, string) bool { return shouldRetry })
			fakeResultSet.FailedReturns(failed)
			fakeResultSet.TotalReturns(failed + succeeded)

			Expect(p.Summary()).To(Equal(
				fmt.Sprintf("FAILED (%s): %d failed attempts to %s exceeded the threshold of %d allowed failures (Total attempts: %d)",
					fakeBaseMeasurement.Name(),
					failed,
					fakeBaseMeasurement.SummaryPhrase(),
					allowedFailures,
					failed+succeeded,
				)))
		})
	})
})
