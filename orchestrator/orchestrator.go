// +build !windows

package orchestrator

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/measurement"
	uuid "github.com/satori/go.uuid"
)

//go:generate counterfeiter . Orchestrator
type Orchestrator interface {
	Setup(cmdRunner.CmdRunner, cfCmdGenerator.CfCmdGenerator) error
	Run(bool) (int, error)
	TearDown(cmdRunner.CmdRunner, cfCmdGenerator.CfCmdGenerator) error
}

type orchestrator struct {
	logger              *log.Logger
	whileConfig         []*config.Command
	workflow            cfWorkflow.CfWorkflow
	whileCommandsRunner cmdRunner.CmdRunner
	measurements        []measurement.Measurement
}

func New(whileConfig []*config.Command, logger *log.Logger, workflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner, measurements []measurement.Measurement) Orchestrator {
	return &orchestrator{
		logger:              logger,
		whileConfig:         whileConfig,
		workflow:            workflow,
		whileCommandsRunner: runner,
		measurements:        measurements,
	}
}

func (o *orchestrator) Setup(runner cmdRunner.CmdRunner, ccg cfCmdGenerator.CfCmdGenerator) error {
	serviceName := fmt.Sprintf("uptimer-srv-%s", uuid.NewV4().String())

	cmds := o.workflow.Setup(ccg)
	cmds = append(cmds, o.workflow.Push(ccg)...)
	cmds = append(cmds, o.workflow.CreateAndBindSyslogDrainService(ccg, serviceName)...)

	return runner.RunInSequence(cmds...)
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
	err := o.whileCommandsRunner.RunInSequence(o.createWhileCmds()...)
	if err != nil {
		exitCode = getExitCodeFromErr(err)
	}
	o.logger.Println("Finished running commands")

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

func (o *orchestrator) TearDown(runner cmdRunner.CmdRunner, ccg cfCmdGenerator.CfCmdGenerator) error {
	return runner.RunInSequence(o.workflow.TearDown(ccg)...)
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
