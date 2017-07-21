package measurement_test

import (
	"bytes"
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

var _ = Describe("RecentLogs", func() {
	var (
		commands             []cmdStartWaiter.CmdStartWaiter
		logger               *log.Logger
		logBuf               *bytes.Buffer
		fakeResultSet        *measurementfakes.FakeResultSet
		fakeAppLogValidator  *appLogValidatorfakes.FakeAppLogValidator
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		rlm BaseMeasurement
	)

	BeforeEach(func() {
		logBuf = bytes.NewBuffer([]byte{})
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		fakeAppLogValidator = &appLogValidatorfakes.FakeAppLogValidator{}
		fakeAppLogValidator.IsNewerReturns(true, nil)

		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}
		logger = log.New(logBuf, "", 0)
		fakeResultSet = &measurementfakes.FakeResultSet{}

		rlm = NewRecentLogs(fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf, fakeAppLogValidator)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(rlm.Name()).To(Equal("Recent logs fetching"))
		})
	})

	Describe("PerformMeasurement", func() {
		It("runs the generated recent logs commands", func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}
			rlm.PerformMeasurement()

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("foo"),
					exec.Command("bar"),
				},
			))
		})

		It("records the commands that run without an error as success", func() {
			_, res := rlm.PerformMeasurement()

			Expect(res).To(BeTrue())
		})

		It("records failure when the app logs are not in order", func() {
			fakeAppLogValidator.IsNewerReturns(false, nil)

			_, res := rlm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("records failure when the app log validator returns an error", func() {
			fakeAppLogValidator.IsNewerReturns(true, fmt.Errorf("oh totally bad news"))

			_, res := rlm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			_, res := rlm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("returns both stdout and stderr when there is an error running the command", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			msg, _ := rlm.PerformMeasurement()

			Expect(msg).To(Equal("\x1b[31mFAILURE(Recent logs fetching): errrrrrooooorrrr\x1b[0m\nstdout:\nheyyy guys\nstderr:\nwhaaats happening?\n\n"))
		})

		It("returns both stdout and stderr when the log validator fails", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, nil)

			msg, _ := rlm.PerformMeasurement()

			Expect(msg).To(Equal("\x1b[31mFAILURE(Recent logs fetching): App log fetched was not newer than previous app log fetched\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("returns both stdout and stderr when the log validator returns an error", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, fmt.Errorf("we don't need no stinking numbers"))

			msg, _ := rlm.PerformMeasurement()

			Expect(msg).To(Equal("\x1b[31mFAILURE(Recent logs fetching): App log validation failed with: we don't need no stinking numbers\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")
			errBuf.WriteString("that's some standard error")

			rlm.PerformMeasurement()

			Expect(outBuf.Len()).To(Equal(0))
			Expect(outBuf.Len()).To(Equal(0))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			fakeResultSet.FailedReturns(0)

			rlm.PerformMeasurement()

			Expect(rlm.Failed(fakeResultSet)).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			fakeResultSet.FailedReturns(1)

			rlm.PerformMeasurement()

			Expect(rlm.Failed(fakeResultSet)).To(BeTrue())
		})
	})
})
