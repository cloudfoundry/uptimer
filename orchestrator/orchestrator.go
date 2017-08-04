package orchestrator

import (
	"log"
	"os/exec"
	"strings"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
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

	var exitCode int
	exitCode, err := o.runWhileCommands()
	if err != nil {
		return exitCode, err
	}

	if performMeasurements {
		for _, m := range o.measurements {
			o.logger.Printf("Stopping measurement: %s\n", m.Name())
			m.Stop()
		}

		o.logger.Println("Measurement summaries:")
		for _, m := range o.measurements {
			if m.Failed() {
				exitCode = 1
				o.logger.Printf("\x1b[31m%s\x1b[0m\n", m.Summary())

			} else {
				o.logger.Printf("\x1b[32m%s\x1b[0m\n", m.Summary())
			}

		}
	}

	return exitCode, nil
}

func (o *orchestrator) TearDown(ccg cfCmdGenerator.CfCmdGenerator) error {
	return o.runner.RunInSequence(o.workflow.TearDown(ccg)...)
}

func (o *orchestrator) runWhileCommands() (int, error) {
	for _, cfg := range o.whileConfig {
		cmd := exec.Command(cfg.Command, cfg.CommandArgs...)
		o.logger.Printf("Running command: `%s %s`\n", o.whileConfig[0].Command, strings.Join(o.whileConfig[0].CommandArgs, " "))
		if err := o.runner.Run(cmd); err != nil {
			return 64, err
		}
		o.logger.Println()

		o.logger.Println("Finished running command")
	}
	return 0, nil
}
