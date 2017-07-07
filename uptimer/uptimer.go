package uptimer

import (
	"log"
	"os/exec"
	"strings"

	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"
)

type Uptimer interface {
	Setup() error
	Run() error
	TearDown() error
}

type uptimer struct {
	Config   *config.Config
	Logger   *log.Logger
	Workflow cfWorkflow.CfWorkflow
	Runner   cmdRunner.CmdRunner
}

func New(config *config.Config, logger *log.Logger, workflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner) Uptimer {
	return &uptimer{
		Config:   config,
		Logger:   logger,
		Workflow: workflow,
		Runner:   runner,
	}
}

func (u *uptimer) Setup() error {
	if err := u.Runner.RunInSequence(u.Workflow.Setup()...); err != nil {
		return err
	}

	return nil
}

func (u *uptimer) Run() error {
	u.Logger.Printf("Running command: `%s %s`\n", u.Config.Command, strings.Join(u.Config.CommandArgs, " "))
	cmd := exec.Command(u.Config.Command, u.Config.CommandArgs...)
	if err := u.Runner.Run(cmd); err != nil {
		u.Logger.Fatalln("Failed running command:", err)
	}
	u.Logger.Println("Finished running command")
	return nil
}

func (u *uptimer) TearDown() error {
	if err := u.Runner.RunInSequence(u.Workflow.TearDown()...); err != nil {
		return err
	}

	return nil
}
