package measurement_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/cloudfoundry/uptimer/appLogValidator/appLogValidatorfakes"
	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"
	"github.com/cloudfoundry/uptimer/measurement/measurementfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StreamLogs", func() {
	var (
		commands             []cmdStartWaiter.CmdStartWaiter
		logger               *log.Logger
		ctx                  context.Context
		logBuf               *bytes.Buffer
		fakeResultSet        *measurementfakes.FakeResultSet
		fakeAppLogValidator  *appLogValidatorfakes.FakeAppLogValidator
		fakeCancelFunc       context.CancelFunc
		cancelFuncCallCount  int
		fakeCmdGeneratorFunc func() (context.Context, context.CancelFunc, []cmdStartWaiter.CmdStartWaiter)
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		slm BaseMeasurement
	)

	BeforeEach(func() {
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
		fakeResultSet = &measurementfakes.FakeResultSet{}

		slm = NewStreamLogs(fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf, fakeAppLogValidator)
	})

	Describe("Name", func() {
		It("returns the name of the measurement", func() {
			Expect(slm.Name()).To(Equal("Streaming logs"))
		})
	})

	Describe("PerformMeasurement", func() {
		It("runs the generated stream logs commands", func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}
			slm.PerformMeasurement(logger)

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

		It("records the commands that run without an error as success", func() {
			res := slm.PerformMeasurement(logger)

			Expect(res).To(BeTrue())
		})

		It("records failure when the app logs are not in order", func() {
			fakeAppLogValidator.IsNewerReturns(false, nil)

			res := slm.PerformMeasurement(logger)

			Expect(res).To(BeFalse())
		})

		It("records failure when the app log validator returns an error", func() {
			fakeAppLogValidator.IsNewerReturns(true, fmt.Errorf("oh totally bad news"))

			res := slm.PerformMeasurement(logger)

			Expect(res).To(BeFalse())
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))

			res := slm.PerformMeasurement(logger)

			Expect(res).To(BeFalse())
		})

		It("calls the cancelfunc when the command does not fail", func() {
			slm.PerformMeasurement(logger)

			Expect(cancelFuncCallCount).To(Equal(1))
		})

		It("calls the cancelfunc when the command fails", func() {
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))

			slm.PerformMeasurement(logger)

			Expect(cancelFuncCallCount).To(Equal(1))
		})

		It("logs both stdout and stderr when there is an error running the command", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("errrrrrooooorrrr"))

			slm.PerformMeasurement(logger)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): errrrrrooooorrrr\x1b[0m\nstdout:\nheyyy guys\nstderr:\nwhaaats happening?\n\n"))
		})

		It("logs both stdout and stderr when the log validator fails", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, nil)

			slm.PerformMeasurement(logger)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): App log fetched was not newer than previous app log fetched\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("logs both stdout and stderr when the log validator returns an error", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, fmt.Errorf("we don't need no stinking numbers"))

			slm.PerformMeasurement(logger)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): App log validation failed with: we don't need no stinking numbers\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")

			slm.PerformMeasurement(logger)

			outBuf.WriteString("first failure")
			errBuf.WriteString("that's some standard error")
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("e 1"))
			slm.PerformMeasurement(logger)

			outBuf.WriteString("second failure")
			errBuf.WriteString("err-body in the club")
			fakeCommandRunner.RunInSequenceWithContextReturns(fmt.Errorf("e 2"))
			slm.PerformMeasurement(logger)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Streaming logs): e 1\x1b[0m\nstdout:\nfirst failure\nstderr:\nthat's some standard error\n\n\x1b[31mFAILURE(Streaming logs): e 2\x1b[0m\nstdout:\nsecond failure\nstderr:\nerr-body in the club\n\n"))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			fakeResultSet.FailedReturns(0)

			slm.PerformMeasurement(logger)

			Expect(slm.Failed(fakeResultSet)).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			fakeResultSet.FailedReturns(1)

			slm.PerformMeasurement(logger)

			Expect(slm.Failed(fakeResultSet)).To(BeTrue())
		})
	})
})
