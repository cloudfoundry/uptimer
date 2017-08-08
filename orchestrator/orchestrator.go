// +build !windows

package orchestrator

import (
	"log"
	"os/exec"
	"syscall"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/measurement"
)

//go:generate counterfeiter . Orchestrator
type Orchestrator interface {
	Setup(cfCmdGenerator.CfCmdGenerator) error
	Run(bool) (int, error)
	TearDown(cfCmdGenerator.CfCmdGenerator) error
}

type orchestrator struct {
	logger       *log.Logger
	whileConfig  []*config.CommandConfig
	workflow     cfWorkflow.CfWorkflow
	runner       cmdRunner.CmdRunner
	measurements []measurement.Measurement
}

func New(whileConfig []*config.CommandConfig, logger *log.Logger, workflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner, measurements []measurement.Measurement) Orchestrator {
	return &orchestrator{
		logger:       logger,
		whileConfig:  whileConfig,
		workflow:     workflow,
		runner:       runner,
		measurements: measurements,
	}
}

func (o *orchestrator) Setup(ccg cfCmdGenerator.CfCmdGenerator) error {
	return o.runner.RunInSequence(append(o.workflow.Setup(ccg), o.workflow.Push(ccg)...)...)
}

func (o *orchestrator) Run(performMeasurements bool) (int, error) {
	if !performMeasurements {
		o.logger.Println("*****NOT PERFORMING ANY MEASUREMENTS*****")
	}

	if performMeasurements {
		for _, m := range o.measurements {
			o.logger.Printf("Starting measurement: %s\n", m.Name())
			go m.Start()
		}
	}

	o.logger.Println("Running commands...")
	exitCode := 0
	err := o.runner.RunInSequence(o.createWhileCmds()...)
	if err != nil {
		exitCode = getExitCodeFromErr(err)
	}
	o.logger.Printf("\nFinished running commands\n")

	if performMeasurements {
		for _, m := range o.measurements {
			o.logger.Printf("Stopping measurement: %s\n", m.Name())
			m.Stop()
		}

		o.logger.Println("Measurement summaries:")
		for _, m := range o.measurements {
			if m.Failed() {
				if exitCode == 0 {
					exitCode = 64
				}
				o.logger.Printf("\x1b[31m%s\x1b[0m\n", m.Summary())

			} else {
				o.logger.Printf("\x1b[32m%s\x1b[0m\n", m.Summary())
			}
		}
	}

	return exitCode, err
}

func (o *orchestrator) TearDown(ccg cfCmdGenerator.CfCmdGenerator) error {
	return o.runner.RunInSequence(o.workflow.TearDown(ccg)...)
}

type Syser interface {
	Sys() interface{}
}

func (o *orchestrator) createWhileCmds() []cmdStartWaiter.CmdStartWaiter {
	cmds := []cmdStartWaiter.CmdStartWaiter{}
	for _, cfg := range o.whileConfig {
		cmds = append(cmds, exec.Command(cfg.Command, cfg.CommandArgs...))
	}

	return cmds
}

func getExitCodeFromErr(err error) int {
	if _, ok := err.(Syser); ok {
		errorCode := err.(Syser).Sys().(syscall.WaitStatus).ExitStatus()
		return errorCode
	}

	return -1
}
