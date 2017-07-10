package orchestrator

import (
	"log"
	"os/exec"
	"strings"

	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
)

type Orchestrator interface {
	Setup() error
	Run() error
	TearDown() error
}

type orchestrator struct {
	WhileConfig *config.CommandConfig
	Logger      *log.Logger
	Workflow    cfWorkflow.CfWorkflow
	Runner      cmdRunner.CmdRunner
}

func New(whileConfig *config.CommandConfig, logger *log.Logger, workflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner) Orchestrator {
	return &orchestrator{
		WhileConfig: whileConfig,
		Logger:      logger,
		Workflow:    workflow,
		Runner:      runner,
	}
}

func (u *orchestrator) Setup() error {
	if err := u.Runner.RunInSequence(u.Workflow.Setup()...); err != nil {
		return err
	}

	return nil
}

func (u *orchestrator) Run() error {
	u.Logger.Printf("Running command: `%s %s`\n", u.WhileConfig.Command, strings.Join(u.WhileConfig.CommandArgs, " "))
	cmd := exec.Command(u.WhileConfig.Command, u.WhileConfig.CommandArgs...)
	if err := u.Runner.Run(cmd); err != nil {
		u.Logger.Fatalln("Failed running command:", err)
	}
	u.Logger.Println("Finished running command")
	return nil
}

func (u *orchestrator) TearDown() error {
	if err := u.Runner.RunInSequence(u.Workflow.TearDown()...); err != nil {
		return err
	}

	return nil
}
