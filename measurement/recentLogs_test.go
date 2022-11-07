package measurement_test

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/uptimer/appLogValidator/appLogValidatorfakes"
	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecentLogs", func() {
	var (
		commands             []cmdStartWaiter.CmdStartWaiter
		fakeAppLogValidator  *appLogValidatorfakes.FakeAppLogValidator
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		rlm BaseMeasurement
	)

	BeforeEach(func() {
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		fakeAppLogValidator = &appLogValidatorfakes.FakeAppLogValidator{}
		fakeAppLogValidator.IsNewerReturns(true, nil)

		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}

		rlm = NewRecentLogs(fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf, fakeAppLogValidator)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(rlm.Name()).To(Equal("Recent logs"))
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
			_, _, _, res := rlm.PerformMeasurement()

			Expect(res).To(BeTrue())
		})

		It("records failure when the app logs are not in order", func() {
			fakeAppLogValidator.IsNewerReturns(false, nil)

			_, _, _, res := rlm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("records failure when the app log validator returns an error", func() {
			fakeAppLogValidator.IsNewerReturns(true, fmt.Errorf("oh totally bad news"))

			_, _, _, res := rlm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			_, _, _, res := rlm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("returns both stdout and stderr when there is an error running the command", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			msg, stdOut, stdErr, _ := rlm.PerformMeasurement()

			Expect(msg).To(Equal("errrrrrooooorrrr"))
			Expect(stdOut).To(Equal("heyyy guys"))
			Expect(stdErr).To(Equal("whaaats happening?"))
		})

		It("returns both stdout and stderr when the log validator fails", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, nil)

			msg, stdOut, stdErr, _ := rlm.PerformMeasurement()

			Expect(msg).To(Equal("App log fetched was not newer than previous app log fetched"))
			Expect(stdOut).To(Equal("yo yo"))
			Expect(stdErr).To(Equal("howayah?"))
		})

		It("returns both stdout and stderr when the log validator returns an error", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, fmt.Errorf("we don't need no stinking numbers"))

			msg, stdOut, stdErr, _ := rlm.PerformMeasurement()

			Expect(msg).To(Equal("App log validation failed with: we don't need no stinking numbers"))
			Expect(stdOut).To(Equal("yo yo"))
			Expect(stdErr).To(Equal("howayah?"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")
			errBuf.WriteString("that's some standard error")

			rlm.PerformMeasurement()

			Expect(outBuf.Len()).To(Equal(0))
			Expect(outBuf.Len()).To(Equal(0))
		})
	})
})
