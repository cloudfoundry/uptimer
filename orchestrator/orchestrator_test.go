package orchestrator_test

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/fakes"
	. "github.com/cloudfoundry/uptimer/orchestrator"

	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/measurement"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Orchestrator", func() {
	var (
		wcfg             *config.CommandConfig
		wcfg1            *config.CommandConfig
		logBuf           *bytes.Buffer
		logger           *log.Logger
		fakeWorkflow     *fakes.FakeCfWorkflow
		fakeRunner       *fakes.FakeCmdRunner
		fakeMeasurement1 *fakes.FakeMeasurement
		fakeMeasurement2 *fakes.FakeMeasurement

		orc Orchestrator
	)

	BeforeEach(func() {
		wcfg = &config.CommandConfig{
			Command:     "sleep",
			CommandArgs: []string{"10"},
		}
		wcfg1 = &config.CommandConfig{
			Command:     "sleep",
			CommandArgs: []string{"15"},
		}
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)
		fakeWorkflow = &fakes.FakeCfWorkflow{}
		fakeRunner = &fakes.FakeCmdRunner{}
		fakeMeasurement1 = &fakes.FakeMeasurement{}
		fakeMeasurement1.NameReturns("name1")
		fakeMeasurement2 = &fakes.FakeMeasurement{}
		fakeMeasurement2.NameReturns("name2")

		orc = New([]*config.CommandConfig{wcfg, wcfg1}, logger, fakeWorkflow, fakeRunner, []measurement.Measurement{fakeMeasurement1, fakeMeasurement2})
	})

	Describe("Setup", func() {
		It("calls workflow to get setup and push stuff and runs it", func() {
			fakeWorkflow.SetupReturns(
				[]cmdStartWaiter.CmdStartWaiter{
					cmdStartWaiter.New(exec.Command("ls")),
					cmdStartWaiter.New(exec.Command("whoami")),
				},
			)
			fakeWorkflow.PushReturns(
				[]cmdStartWaiter.CmdStartWaiter{
					cmdStartWaiter.New(exec.Command("push", "an", "app")),
				},
			)

			err := orc.Setup()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeWorkflow.SetupCallCount()).To(Equal(1))
			Expect(fakeWorkflow.PushCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					cmdStartWaiter.New(exec.Command("ls")),
					cmdStartWaiter.New(exec.Command("whoami")),
					cmdStartWaiter.New(exec.Command("push", "an", "app")),
				},
			))
		})

		It("Returns an error if runner returns an error", func() {
			fakeRunner.RunInSequenceReturns(fmt.Errorf("uh oh"))

			err := orc.Setup()

			Expect(err).To(MatchError("uh oh"))
		})
	})

	Describe("Run", func() {
		Context("when all the while commands pass", func() {
			It("runs all the given while commands", func() {
				_, err := orc.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeRunner.RunCallCount()).To(Equal(2))
				Expect(fakeRunner.RunArgsForCall(0)).To(Equal(cmdStartWaiter.New(exec.Command("sleep", "10"))))
				Expect(fakeRunner.RunArgsForCall(1)).To(Equal(cmdStartWaiter.New(exec.Command("sleep", "15"))))
			})
		})

		Context("when the first while command fails", func() {
			It("returns an error with exit code 64 and doesn't run the second command", func() {
				fakeRunner.RunReturns(fmt.Errorf("oh boy"))

				exitCode, err := orc.Run()

				Expect(fakeRunner.RunArgsForCall(0)).To(Equal(cmdStartWaiter.New(exec.Command("sleep", "10"))))
				Expect(fakeRunner.RunCallCount()).To(Equal(1))
				Expect(exitCode).To(Equal(64))
				Expect(err).To(MatchError("oh boy"))
			})
		})

		Context("when the second command fails", func() {
			It("returns an error with exit code 64", func() {
				fakeRunner.RunReturnsOnCall(1, fmt.Errorf("oh boy"))

				exitCode, err := orc.Run()

				Expect(fakeRunner.RunArgsForCall(0)).To(Equal(cmdStartWaiter.New(exec.Command("sleep", "10"))))
				Expect(fakeRunner.RunArgsForCall(1)).To(Equal(cmdStartWaiter.New(exec.Command("sleep", "15"))))
				Expect(fakeRunner.RunCallCount()).To(Equal(2))
				Expect(exitCode).To(Equal(64))
				Expect(err).To(MatchError("oh boy"))
			})
		})

		It("prints a list of all measurements", func() {
			orc.Run()

			Expect(logBuf.String()).To(ContainSubstring("Starting measurement: name1"))
			Expect(logBuf.String()).To(ContainSubstring("Starting measurement: name2"))
		})

		It("starts all the measurements once", func() {
			orc.Run()

			Eventually(fakeMeasurement1.StartCallCount, 3*time.Second).Should(Equal(1))
			Eventually(fakeMeasurement2.StartCallCount, 3*time.Second).Should(Equal(1))
		})

		It("stops all the measurements once", func() {
			orc.Run()

			Expect(fakeMeasurement1.StopCallCount()).To(Equal(1))
			Expect(fakeMeasurement2.StopCallCount()).To(Equal(1))
		})

		It("gathers the sumaries and prints them all with a header", func() {
			fakeMeasurement1.SummaryReturns("summary1")
			fakeMeasurement2.SummaryReturns("summary2")

			orc.Run()

			Expect(fakeMeasurement1.SummaryCallCount()).To(Equal(1))
			Expect(fakeMeasurement2.SummaryCallCount()).To(Equal(1))
			Expect(logBuf.String()).To(ContainSubstring("Measurement summaries:"))
			Expect(logBuf.String()).To(ContainSubstring("summary1"))
			Expect(logBuf.String()).To(ContainSubstring("summary2"))
		})

		It("returns an exit code of 0 when all measurements succeed", func() {
			ec, _ := orc.Run()

			Expect(ec).To(Equal(0))
		})

		It("returns an exit code of 1 when any measurement fails", func() {
			fakeMeasurement1.FailedReturns(true)

			ec, _ := orc.Run()

			Expect(ec).To(Equal(1))
		})
	})

	Describe("TearDown", func() {
		It("calls workflow to get teardown stuff and runs it", func() {
			fakeWorkflow.TearDownReturns(
				[]cmdStartWaiter.CmdStartWaiter{
					cmdStartWaiter.New(exec.Command("ls")),
					cmdStartWaiter.New(exec.Command("whoami")),
				},
			)

			err := orc.TearDown()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeWorkflow.TearDownCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					cmdStartWaiter.New(exec.Command("ls")),
					cmdStartWaiter.New(exec.Command("whoami")),
				},
			))
		})

		It("Returns an error if runner returns an error", func() {
			fakeRunner.RunInSequenceReturns(fmt.Errorf("uh oh"))

			err := orc.TearDown()

			Expect(err).To(MatchError("uh oh"))
		})
	})
})
