package orchestrator

import (
	"log"
	"os/exec"
	"strings"

	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/measurement"
)

type Orchestrator interface {
	Setup() error
	Run() error
	TearDown() error
}

type orchestrator struct {
	Logger       *log.Logger
	WhileConfig  *config.CommandConfig
	Workflow     cfWorkflow.CfWorkflow
	Runner       cmdRunner.CmdRunner
	Measurements []measurement.Measurement
}

func New(whileConfig *config.CommandConfig, logger *log.Logger, workflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner, measurements []measurement.Measurement) Orchestrator {
	return &orchestrator{
		Logger:       logger,
		WhileConfig:  whileConfig,
		Workflow:     workflow,
		Runner:       runner,
		Measurements: measurements,
	}
}

func (o *orchestrator) Setup() error {
	if err := o.Runner.RunInSequence(o.Workflow.Setup()...); err != nil {
		return err
	}

	return nil
}

func (o *orchestrator) Run() error {
	for _, m := range o.Measurements {
		o.Logger.Printf("Starting measurement: %s\n", m.Name())
		go m.Start()
	}

	cmd := exec.Command(o.WhileConfig.Command, o.WhileConfig.CommandArgs...)
	o.Logger.Printf("Running command: `%s %s`\n", o.WhileConfig.Command, strings.Join(o.WhileConfig.CommandArgs, " "))
	if err := o.Runner.Run(cmd); err != nil {
		return err
	}
	o.Logger.Println("Finished running command")

	o.Logger.Println("Measurement summaries:")
	for _, m := range o.Measurements {
		m.Stop()
		o.Logger.Println(m.Summary())
	}

	return nil
}

func (o *orchestrator) TearDown() error {
	if err := o.Runner.RunInSequence(o.Workflow.TearDown()...); err != nil {
		return err
	}

	return nil
}
