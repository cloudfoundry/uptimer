package measurement_test

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pushability", func() {
	var (
		commands             []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		pm BaseMeasurement
	)

	BeforeEach(func() {
		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		pm = NewAppPushability(fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(pm.Name()).To(Equal("App pushability"))
		})
	})

	Describe("PerformMeasurement", func() {
		It("runs the generated app push and delete", func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}

			pm.PerformMeasurement()

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("foo"),
					exec.Command("bar"),
				},
			))
		})

		It("records the commands that run without an error as success", func() {
			_, _, _, res := pm.PerformMeasurement()

			Expect(res).To(BeTrue())
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			_, _, _, res := pm.PerformMeasurement()

			Expect(res).To(BeFalse())
		})

		It("returns both stdout and stderr when there is an error", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			msg, stdOut, stdErr, _ := pm.PerformMeasurement()

			Expect(msg).To(Equal("errrrrrooooorrrr"))
			Expect(stdOut).To(Equal("heyyy guys"))
			Expect(stdErr).To(Equal("whaaats happening?"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")
			errBuf.WriteString("that's some standard error")

			pm.PerformMeasurement()

			Expect(outBuf.Len()).To(Equal(0))
			Expect(errBuf.Len()).To(Equal(0))
		})
	})
})
