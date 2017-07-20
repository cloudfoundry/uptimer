package measurement_test

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"

	"github.com/cloudfoundry/uptimer/measurement/measurementfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pushability", func() {
	var (
		commands             []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		fakeResultSet        *measurementfakes.FakeResultSet
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer
		logBuf               *bytes.Buffer
		logger               *log.Logger

		pm BaseMeasurement
	)

	BeforeEach(func() {
		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}
		fakeResultSet = &measurementfakes.FakeResultSet{}
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)

		pm = NewPushability(fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf)
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

			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("foo"),
					exec.Command("bar"),
				},
			))
		})

		It("records the commands that run without an error as success", func() {
			pm.PerformMeasurement(logger, fakeResultSet)
			pm.PerformMeasurement(logger, fakeResultSet)
			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(fakeResultSet.RecordSuccessCallCount()).To(Equal(3))
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			pm.PerformMeasurement(logger, fakeResultSet)
			pm.PerformMeasurement(logger, fakeResultSet)
			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(fakeResultSet.RecordFailureCallCount()).To(Equal(3))
		})

		It("logs both stdout and stderr when there is an error", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(App pushability): errrrrrooooorrrr\x1b[0m\nstdout:\nheyyy guys\nstderr:\nwhaaats happening?\n\n"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")
			pm.PerformMeasurement(logger, fakeResultSet)

			outBuf.WriteString("first failure")
			errBuf.WriteString("that's some standard error")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("e 1"))
			pm.PerformMeasurement(logger, fakeResultSet)

			outBuf.WriteString("second failure")
			errBuf.WriteString("err-body in the club")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("e 2"))
			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(App pushability): e 1\x1b[0m\nstdout:\nfirst failure\nstderr:\nthat's some standard error\n\n\x1b[31mFAILURE(App pushability): e 2\x1b[0m\nstdout:\nsecond failure\nstderr:\nerr-body in the club\n\n"))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			fakeResultSet.FailedReturns(0)

			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(pm.Failed(fakeResultSet)).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			fakeResultSet.FailedReturns(1)

			pm.PerformMeasurement(logger, fakeResultSet)

			Expect(pm.Failed(fakeResultSet)).To(BeTrue())
		})
	})
})
