package orchestrator_test

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow/cfWorkflowfakes"
	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/measurement/measurementfakes"
	. "github.com/cloudfoundry/uptimer/orchestrator"

	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/measurement"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Orchestrator", func() {
	var (
		fakeCommand1     *config.Command
		fakeCommand2     *config.Command
		logBuf           *bytes.Buffer
		logger           *log.Logger
		fakeWorkflow     *cfWorkflowfakes.FakeCfWorkflow
		fakeRunner       *cmdRunnerfakes.FakeCmdRunner
		fakeMeasurement1 *measurementfakes.FakeMeasurement
		fakeMeasurement2 *measurementfakes.FakeMeasurement
		ccg              cfCmdGenerator.CfCmdGenerator

		orc Orchestrator
	)

	BeforeEach(func() {
		fakeCommand1 = &config.Command{
			Command:     "sleep",
			CommandArgs: []string{"10"},
		}
		fakeCommand2 = &config.Command{
			Command:     "sleep",
			CommandArgs: []string{"15"},
		}
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)
		fakeWorkflow = &cfWorkflowfakes.FakeCfWorkflow{}
		fakeRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeMeasurement1 = &measurementfakes.FakeMeasurement{}
		fakeMeasurement1.NameReturns("name1")
		fakeMeasurement2 = &measurementfakes.FakeMeasurement{}
		fakeMeasurement2.NameReturns("name2")
		ccg = cfCmdGenerator.New("/cfhome")

		orc = New([]*config.Command{fakeCommand1, fakeCommand2}, logger, fakeWorkflow, fakeRunner, []measurement.Measurement{fakeMeasurement1, fakeMeasurement2})
	})

	Describe("Setup", func() {
		It("calls workflow to get setup and push stuff and runs it", func() {
			fakeWorkflow.SetupReturns(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			)
			fakeWorkflow.PushReturns(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("push", "an", "app"),
				},
			)

			err := orc.Setup(fakeRunner, ccg)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeWorkflow.SetupCallCount()).To(Equal(1))
			Expect(fakeWorkflow.SetupArgsForCall(0)).To(Equal(ccg))
			Expect(fakeWorkflow.PushCallCount()).To(Equal(1))
			Expect(fakeWorkflow.PushArgsForCall(0)).To(Equal(ccg))
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
					exec.Command("push", "an", "app"),
				},
			))
		})

		It("Returns an error if runner returns an error", func() {
			fakeRunner.RunInSequenceReturns(fmt.Errorf("uh oh"))

			err := orc.Setup(fakeRunner, ccg)

			Expect(err).To(MatchError("uh oh"))
		})
	})

	Describe("Run", func() {
		It("runs all the given while commands", func() {
			_, err := orc.Run(true)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("sleep", "10"),
					exec.Command("sleep", "15"),
				},
			))
		})

		Context("when a while command fails", func() {
			It("returns an error with exit code of the failed while command", func() {
				exitError := &fakeSyser{
					WS: 512,
				}
				fakeRunner.RunInSequenceReturns(exitError)

				exitCode, _ := orc.Run(true)

				Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						exec.Command("sleep", "10"),
						exec.Command("sleep", "15"),
					},
				))
				Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
				Expect(exitCode).To(Equal(2))
			})

			It("returns an error with exit code of -1 if the failed while command's error is not an exiterror", func() {
				fakeRunner.RunInSequenceReturns(fmt.Errorf("hey dude"))

				exitCode, err := orc.Run(true)

				Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						exec.Command("sleep", "10"),
						exec.Command("sleep", "15"),
					},
				))
				Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
				Expect(exitCode).To(Equal(-1))
				Expect(err).To(MatchError("hey dude"))
			})
		})

		Context("When performMeasurements is false", func() {
			var (
				performMeasurements bool
			)

			BeforeEach(func() {
				performMeasurements = false
			})

			It("Prints a message about not running measurements", func() {
				orc.Run(performMeasurements)

				Expect(logBuf.String()).To(ContainSubstring("*****NOT PERFORMING ANY MEASUREMENTS*****"))
			})

			It("does not print a list of all measurements starting", func() {
				orc.Run(performMeasurements)

				Expect(logBuf.String()).NotTo(ContainSubstring("Starting measurement: name1"))
				Expect(logBuf.String()).NotTo(ContainSubstring("Starting measurement: name2"))
			})

			It("starts no measurements", func() {
				orc.Run(performMeasurements)

				Eventually(fakeMeasurement1.StartCallCount, 3*time.Second).Should(Equal(0))
				Eventually(fakeMeasurement2.StartCallCount, 3*time.Second).Should(Equal(0))
			})

			It("stops no measurements", func() {
				orc.Run(performMeasurements)

				Expect(fakeMeasurement1.StopCallCount()).To(Equal(0))
				Expect(fakeMeasurement2.StopCallCount()).To(Equal(0))
			})

			It("does not print a list of all measurements stopping", func() {
				orc.Run(performMeasurements)

				Expect(logBuf.String()).NotTo(ContainSubstring("Stopping measurement: name1"))
				Expect(logBuf.String()).NotTo(ContainSubstring("Stopping measurement: name2"))
			})

			It("does not gather the sumaries and print them all with a header", func() {
				fakeMeasurement1.SummaryReturns("summary1")
				fakeMeasurement1.FailedReturns(true)
				fakeMeasurement2.SummaryReturns("summary2")

				orc.Run(performMeasurements)

				Expect(fakeMeasurement1.SummaryCallCount()).To(Equal(0))
				Expect(fakeMeasurement2.SummaryCallCount()).To(Equal(0))
				Expect(logBuf.String()).NotTo(ContainSubstring("Measurement summaries:"))
				Expect(logBuf.String()).NotTo(ContainSubstring("\x1b[31msummary1\x1b[0m"))
				Expect(logBuf.String()).NotTo(ContainSubstring("\x1b[32msummary2\x1b[0m"))
			})

			It("returns an exit code of 0 when the while commands succeed", func() {
				ec, _ := orc.Run(performMeasurements)

				Expect(ec).To(Equal(0))
			})

			It("returns the exit code of the failed while command when a while command fails", func() {
				exitError := &fakeSyser{
					WS: 512,
				}
				fakeRunner.RunInSequenceReturns(exitError)

				ec, _ := orc.Run(performMeasurements)

				Expect(ec).To(Equal(2))
			})
		})

		Context("When peformMeasurements is true", func() {
			var (
				performMeasurements bool
			)

			BeforeEach(func() {
				performMeasurements = true
			})

			It("prints a list of all measurements starting", func() {
				orc.Run(performMeasurements)

				Expect(logBuf.String()).To(ContainSubstring("Starting measurement: name1"))
				Expect(logBuf.String()).To(ContainSubstring("Starting measurement: name2"))
			})

			It("starts all the measurements once", func() {
				orc.Run(performMeasurements)

				Eventually(fakeMeasurement1.StartCallCount, 3*time.Second).Should(Equal(1))
				Eventually(fakeMeasurement2.StartCallCount, 3*time.Second).Should(Equal(1))
			})

			It("stops all the measurements once", func() {
				orc.Run(performMeasurements)

				Expect(fakeMeasurement1.StopCallCount()).To(Equal(1))
				Expect(fakeMeasurement2.StopCallCount()).To(Equal(1))
			})

			It("prints a list of all measurements stopping", func() {
				orc.Run(performMeasurements)

				Expect(logBuf.String()).To(ContainSubstring("Stopping measurement: name1"))
				Expect(logBuf.String()).To(ContainSubstring("Stopping measurement: name2"))
			})

			It("gathers the sumaries and prints them all with a header", func() {
				fakeMeasurement1.SummaryReturns("summary1")
				fakeMeasurement1.FailedReturns(true)
				fakeMeasurement2.SummaryReturns("summary2")

				orc.Run(performMeasurements)

				Expect(fakeMeasurement1.SummaryCallCount()).To(Equal(1))
				Expect(fakeMeasurement2.SummaryCallCount()).To(Equal(1))
				Expect(logBuf.String()).To(ContainSubstring("Measurement summaries:"))
				Expect(logBuf.String()).To(ContainSubstring("\x1b[31msummary1\x1b[0m"))
				Expect(logBuf.String()).To(ContainSubstring("\x1b[32msummary2\x1b[0m"))
			})

			It("returns an exit code of 0 when all measurements succeed and the while commands all succeed", func() {
				ec, _ := orc.Run(performMeasurements)

				Expect(ec).To(Equal(0))
			})

			It("returns an exit code of 64 when any measurement fails and the while commands all succeed", func() {
				fakeMeasurement1.FailedReturns(true)

				ec, _ := orc.Run(performMeasurements)

				Expect(ec).To(Equal(64))
			})

			It("returns the failed while command's exit code even when measurements succeed", func() {
				exitError := &fakeSyser{
					WS: 512,
				}
				fakeRunner.RunInSequenceReturns(exitError)

				exitCode, _ := orc.Run(performMeasurements)

				Expect(exitCode).To(Equal(2))
			})

			It("returns the failed while command's exit code even when measurements also fail", func() {
				fakeMeasurement1.FailedReturns(true)
				exitError := &fakeSyser{
					WS: 512,
				}
				fakeRunner.RunInSequenceReturns(exitError)

				exitCode, _ := orc.Run(performMeasurements)

				Expect(exitCode).To(Equal(2))
			})
		})
	})

	Describe("TearDown", func() {
		It("calls workflow to get teardown stuff and runs it", func() {
			fakeWorkflow.TearDownReturns(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			)

			err := orc.TearDown(fakeRunner, ccg)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeWorkflow.TearDownCallCount()).To(Equal(1))
			Expect(fakeWorkflow.TearDownArgsForCall(0)).To(Equal(ccg))
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			))
		})

		It("Returns an error if runner returns an error", func() {
			fakeRunner.RunInSequenceReturns(fmt.Errorf("uh oh"))

			err := orc.TearDown(fakeRunner, ccg)

			Expect(err).To(MatchError("uh oh"))
		})
	})
})

type fakeSyser struct {
	WS syscall.WaitStatus
}

func (f *fakeSyser) Sys() interface{} {
	return f.WS
}

func (f *fakeSyser) Error() string {
	return ""
}
