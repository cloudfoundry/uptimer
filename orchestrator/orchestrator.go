//go:build !windows
// +build !windows

package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"code.cloudfoundry.org/goshims/ioutilshim"

	uuid "github.com/satori/go.uuid"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/measurement"
)

//go:generate counterfeiter . Orchestrator
type Orchestrator interface {
	Setup(cmdRunner.CmdRunner, cfCmdGenerator.CfCmdGenerator, config.OptionalTests) error
	Run(bool, string) (int, error)
	TearDown(cmdRunner.CmdRunner, cfCmdGenerator.CfCmdGenerator) error
}

type orchestrator struct {
	logger              *log.Logger
	whileConfig         []*config.Command
	workflow            cfWorkflow.CfWorkflow
	whileCommandsRunner cmdRunner.CmdRunner
	measurements        []measurement.Measurement
	ioutilshim          ioutilshim.Ioutil
}

type result struct {
	Summaries   []measurement.Summary `json:"summaries"`
	CmdExitCode int                   `json:"commandExitCode"`
}

func New(whileConfig []*config.Command, logger *log.Logger, workflow cfWorkflow.CfWorkflow, runner cmdRunner.CmdRunner, measurements []measurement.Measurement, ioutilShim ioutilshim.Ioutil) Orchestrator {
	return &orchestrator{
		logger:              logger,
		whileConfig:         whileConfig,
		workflow:            workflow,
		whileCommandsRunner: runner,
		measurements:        measurements,
		ioutilshim:          ioutilShim,
	}
}

func (o *orchestrator) Setup(runner cmdRunner.CmdRunner, ccg cfCmdGenerator.CfCmdGenerator, optionalTests config.OptionalTests) error {
	serviceName := fmt.Sprintf("uptimer-srv-%s", uuid.NewV4().String())

	cmds := o.workflow.Setup(ccg)
	cmds = append(cmds, o.workflow.Push(ccg)...)

	if optionalTests.RunAppSyslogAvailability {
		cmds = append(cmds, o.workflow.CreateAndBindSyslogDrainService(ccg, serviceName)...)
	}

	return runner.RunInSequence(cmds...)
}

func (o *orchestrator) Run(performMeasurements bool, resultFilePath string) (int, error) {
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
	commandExitCode := 0
	err := o.whileCommandsRunner.RunInSequence(o.createWhileCmds()...)
	if err != nil {
		exitCode = getExitCodeFromErr(err)
		commandExitCode = exitCode
	}
	o.logger.Println("Finished running commands")

	if performMeasurements {
		for _, m := range o.measurements {
			o.logger.Printf("Stopping measurement: %s\n", m.Name())
			m.Stop()
			o.logger.Printf("Stopped measurement: %s\n", m.Name())
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

		if resultFilePath != "" {
			r := result{}
			for _, m := range o.measurements {
				r.Summaries = append(r.Summaries, m.SummaryData())
			}
			r.CmdExitCode = commandExitCode
			resultJSON, err := json.Marshal(r)
			if err != nil {
				o.logger.Printf("WARN: Failed to serilaize results to json: %s", err.Error())
			}
			err = o.ioutilshim.WriteFile(resultFilePath, resultJSON, os.ModePerm)
			if err != nil {
				o.logger.Printf("WARN: Failed to write result JSON to file: %s", err.Error())
			}
		}
	}

	// Alert user that the While Command succeeded, but we failed in the setup of one or more measurements
	if !performMeasurements && exitCode == 0 {
		exitCode = 70
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
