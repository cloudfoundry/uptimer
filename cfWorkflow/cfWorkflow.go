package cfWorkflow

import "github.com/cloudfoundry/uptimer/cfCmdGenerator"
import "os/exec"

type CfWorkflow interface {
	Setup() []*exec.Cmd
	Cleanup() []*exec.Cmd
}

type cfWorkflow struct {
	ApiUrl            string
	Username          string
	Password          string
	Org               string
	Space             string
	AppName           string
	AppPath           string
	SkipSslValidation bool
	CfCmdGenerator    cfCmdGenerator.CfCmdGenerator
}

func New(
	apiUrl string,
	username string,
	password string,
	org string,
	space string,
	appName string,
	appPath string,
	skipSslValidation bool,
	cfCmdGenerator cfCmdGenerator.CfCmdGenerator,
) CfWorkflow {
	return &cfWorkflow{
		ApiUrl:            apiUrl,
		Username:          username,
		Password:          password,
		Org:               org,
		Space:             space,
		AppName:           appName,
		AppPath:           appPath,
		SkipSslValidation: skipSslValidation,
		CfCmdGenerator:    cfCmdGenerator,
	}
}

func (c *cfWorkflow) Setup() []*exec.Cmd {
	return []*exec.Cmd{
		c.CfCmdGenerator.Api(c.ApiUrl, c.SkipSslValidation),
		c.CfCmdGenerator.Auth(c.Username, c.Password),
		c.CfCmdGenerator.CreateOrg(c.Org),
		c.CfCmdGenerator.CreateSpace(c.Org, c.Space),
		c.CfCmdGenerator.Target(c.Org, c.Space),
		c.CfCmdGenerator.Push(c.AppName, c.AppPath),
	}
}

func (c *cfWorkflow) Cleanup() []*exec.Cmd {
	return []*exec.Cmd{
		c.CfCmdGenerator.DeleteOrg(c.Org),
	}
}
