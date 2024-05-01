package measurement_test

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stats Availability", func() {
	var (
		commands             []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		sm BaseMeasurement
	)

	BeforeEach(func() {
		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		sm = NewStatsAvailability(fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(sm.Name()).To(Equal("Stats availability"))
		})
	})

	Describe("SummaryPhrase", func() {
		It("returns the summary phrase", func() {
			Expect(sm.SummaryPhrase()).To(Equal("retrieve stats for app"))
		})
	})

	Describe("PerformMeasurement", func() {
		BeforeEach(func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}
		})

		It("runs the commands to retrieve the stats for the app", func() {
			sm.PerformMeasurement()

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("foo"),
					exec.Command("bar"),
				},
			))
		})

		It("records the commands that run without an error as success", func() {
			_, _, _, res := sm.PerformMeasurement()
			Expect(res).To(BeTrue())
		})

		Context("when the CLI reports that stats server is unavailable", func() {
			BeforeEach(func() {
				errBuf.WriteString("Stats server temporarily unavailable.")
			})
			It("records the measurement as having failed", func() {
				_, _, _, res := sm.PerformMeasurement()
				Expect(res).To(BeFalse())
			})
		})

		Context("when the commands error", func() {
			BeforeEach(func() {
				fakeCommandRunner.RunInSequenceReturns(errors.New("some error"))
			})

			It("records the measurement as having failed", func() {
				_, _, _, res := sm.PerformMeasurement()
				Expect(res).To(BeFalse())
			})

			It("returns both stdout and stderr", func() {
				outBuf.WriteString("some stdout output")
				errBuf.WriteString("some stderr output")
				msg, stdOut, stdErr, _ := sm.PerformMeasurement()

				Expect(msg).To(Equal("some error"))
				Expect(stdOut).To(Equal("some stdout output"))
				Expect(stdErr).To(Equal("some stderr output"))
			})
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("some stdout output")
			errBuf.WriteString("some stderr output")

			sm.PerformMeasurement()

			Expect(outBuf.Len()).To(Equal(0))
			Expect(errBuf.Len()).To(Equal(0))
		})
	})
})
