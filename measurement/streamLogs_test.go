package measurement_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/appLogValidator/appLogValidatorfakes"
	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StreamLogs", func() {
	var (
		freq                 time.Duration
		mockClock            *clock.Mock
		commands             []cmdStartWaiter.CmdStartWaiter
		logger               *log.Logger
		ctx                  context.Context
		logBuf               *bytes.Buffer
		fakeAppLogValidator  *appLogValidatorfakes.FakeAppLogValidator
		fakeCancelFunc       context.CancelFunc
		cancelFuncCallCount  int
		fakeCmdGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter)
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		slm Measurement
	)

	BeforeEach(func() {
		freq = time.Second
		mockClock = clock.NewMock()
		logBuf = bytes.NewBuffer([]byte{})
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		ctx = context.TODO()
		cancelFuncCallCount = 0
		fakeCancelFunc = func() {
			cancelFuncCallCount++
		}

		fakeAppLogValidator = &appLogValidatorfakes.FakeAppLogValidator{}
		fakeAppLogValidator.IsNewerReturns(true, nil)

		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter) {
			return ctx, fakeCancelFunc, commands
		}
		logger = log.New(logBuf, "", 0)

		slm = NewStreamLogs(logger, freq, mockClock, fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf, fakeAppLogValidator)
	})

	Describe("Name", func() {
		It("returns the name of the measurement", func() {
			Expect(slm.Name()).To(Equal("Streaming logs"))
		})
	})

	Describe("Start", func() {
		AfterEach(func() {
			slm.Stop()
		})

		It("runs the generated stream logs commands", func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}
			err := slm.Start()
			mockClock.Add(freq)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCommandRunner.RunInSequenceWithContextCallCount()).To(BeNumerically(">=", 1))
			actualCtx, actualCmds := fakeCommandRunner.RunInSequenceWithContextArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualCmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("foo"),
					exec.Command("bar"),
				},
			))
		})

		It("runs the stream logs commands with given frequency", func() {
			slm.Start()
			mockClock.Add(3 * freq)

			Expect(fakeCommandRunner.RunInSequenceWithContextCallCount()).To(Equal(4))
		})

		It("records the commands that run without an error as success", func() {
			slm.Start()
			mockClock.Add(3 * freq)

			rs, _ := slm.Results()
			Expect(rs.Successful()).To(Equal(4))
		})

		It("records failure when the app logs are not in order", func() {
			fakeAppLogValidator.IsNewerReturns(false, nil)

			slm.Start()
			mockClock.Add(freq - time.Nanosecond)

			rs, _ := slm.Results()
			Expect(rs.Failed()).To(Equal(1))
		})

		It("records failure when the app log validator returns an error", func() {
			fakeAppLogValidator.IsNewerReturns(true, fmt.Errorf("oh totally bad news"))

			slm.Start()
			mockClock.Add(freq - time.Nanosecond)

			rs, _ := slm.Results()
			Expect(rs.Failed()).To(Equal(1))
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))

			slm.Start()
			mockClock.Add(3 * freq)

			rs, _ := slm.Results()
			Expect(rs.Failed()).To(Equal(4))
		})

		It("records all of the results", func() {
			slm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(3 * freq)

			rs, _ := slm.Results()
			Expect(rs.Successful()).To(Equal(4))
			Expect(rs.Failed()).To(Equal(3))
			Expect(rs.Total()).To(Equal(7))
		})

		It("calls the cancelfunc when the command does not fail", func() {
			slm.Start()
			mockClock.Add(3 * freq)

			Expect(cancelFuncCallCount).To(Equal(4))
		})

		It("calls the cancelfunc when the command fails", func() {
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))
			slm.Start()
			mockClock.Add(3 * freq)

			Expect(cancelFuncCallCount).To(Equal(4))
		})

		It("logs both stdout and stderr when there is an error running the command", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))

			slm.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): errrrrrooooorrrr\x1b[0m\nstdout:\nheyyy guys\nstderr:\nwhaaats happening?\n\n"))
		})

		It("logs both stdout and stderr when the log validator fails", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, nil)

			slm.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): App log fetched was not newer than previous app log fetched\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("logs both stdout and stderr when the log validator returns an error", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, fmt.Errorf("we don't need no stinking numbers"))

			slm.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): App log validation failed with: we don't need no stinking numbers\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")

			slm.Start()
			mockClock.Add(freq - time.Nanosecond)

			outBuf.WriteString("first failure")
			errBuf.WriteString("that's some standard error")
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("e 1"))
			mockClock.Add(freq)

			outBuf.WriteString("second failure")
			errBuf.WriteString("err-body in the club")
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("e 2"))
			mockClock.Add(freq)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): e 1\x1b[0m\nstdout:\nfirst failure\nstderr:\nthat's some standard error\n\n\x1b[31mFAILURE(Streaming logs): e 2\x1b[0m\nstdout:\nsecond failure\nstderr:\nerr-body in the club\n\n"))
		})
	})

	Describe("Stop", func() {
		It("stops the measurement", func() {
			slm.Start()
			mockClock.Add(3 * freq)
			slm.Stop()
			mockClock.Add(3 * freq)

			Expect(fakeCommandRunner.RunInSequenceWithContextCallCount()).To(Equal(4))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			slm.Start()
			mockClock.Add(3 * freq)

			Expect(slm.Failed()).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			slm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(freq)

			Expect(slm.Failed()).To(BeTrue())
		})
	})

	Describe("Summary", func() {
		It("returns a success summary if none failed", func() {
			slm.Start()
			mockClock.Add(3 * freq)
			slm.Stop()

			Expect(slm.Summary()).To(Equal(fmt.Sprintf("SUCCESS(%s): All %d attempts to stream logs succeeded", slm.Name(), 4)))
		})

		It("returns a failed summary if there are failures", func() {
			slm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(3 * freq)
			slm.Stop()

			Expect(slm.Summary()).To(Equal(fmt.Sprintf("FAILED(%s): %d of %d attempts to stream logs failed", slm.Name(), 3, 7)))
		})
	})
})
