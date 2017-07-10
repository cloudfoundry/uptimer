package orchestrator_test

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/fakes"
	. "github.com/cloudfoundry/uptimer/orchestrator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Orchestrator", func() {
	var (
		wcfg         *config.CommandConfig
		logBuf       *bytes.Buffer
		logger       *log.Logger
		fakeWorkflow *fakes.FakeCfWorkflow
		fakeRunner   *fakes.FakeCmdRunner

		uptimer Orchestrator
	)

	BeforeEach(func() {
		wcfg = &config.CommandConfig{}
		logBuf = bytes.NewBuffer([]byte{})
		logger = log.New(logBuf, "", 0)
		fakeWorkflow = &fakes.FakeCfWorkflow{}
		fakeRunner = &fakes.FakeCmdRunner{}

		uptimer = New(wcfg, logger, fakeWorkflow, fakeRunner)
	})

	Describe("Setup", func() {
		It("calls workflow to get setup stuff and runs it", func() {
			fakeWorkflow.SetupReturns(
				[]cmdRunner.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			)

			err := uptimer.Setup()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeWorkflow.SetupCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdRunner.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			))
		})

		It("Returns an error if runner returns an error", func() {
			fakeRunner.RunInSequenceReturns(fmt.Errorf("uh oh"))

			err := uptimer.Setup()

			Expect(err).To(MatchError("uh oh"))
		})
	})

	Describe("Run", func() {
		PIt("runs a measurement and returns a report")
	})

	Describe("TearDown", func() {
		It("calls workflow to get teardown stuff and runs it", func() {
			fakeWorkflow.TearDownReturns(
				[]cmdRunner.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			)

			err := uptimer.TearDown()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeWorkflow.TearDownCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceCallCount()).To(Equal(1))
			Expect(fakeRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdRunner.CmdStartWaiter{
					exec.Command("ls"),
					exec.Command("whoami"),
				},
			))
		})

		It("Returns an error if runner returns an error", func() {
			fakeRunner.RunInSequenceReturns(fmt.Errorf("uh oh"))

			err := uptimer.TearDown()

			Expect(err).To(MatchError("uh oh"))
		})
	})
})
