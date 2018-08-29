package cfWorkflow

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"
)

//go:generate counterfeiter . CfWorkflow
type CfWorkflow interface {
	Org() string
	Space() string
	Quota() string
	AppUrl() string

	Setup(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	Push(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	Delete(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	TearDown(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	RecentLogs(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	StreamLogs(context.Context, cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter

	MapRoute(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	CreateAndBindSyslogDrainService(cfCmdGenerator.CfCmdGenerator, string) []cmdStartWaiter.CmdStartWaiter
}

type cfWorkflow struct {
	cf *config.Cf

	appPath    string
	org        string
	space      string
	quota      string
	appName    string
	appCommand string
}

func (c *cfWorkflow) Org() string {
	return c.org
}

func (c *cfWorkflow) Space() string {
	return c.space
}

func (c *cfWorkflow) Quota() string {
	return c.quota
}

func (c *cfWorkflow) AppUrl() string {
	return fmt.Sprintf("https://%s.%s", c.appName, c.cf.AppDomain)
}

func New(cfConfig *config.Cf, org, space, quota, appName, appPath, appCommand string) CfWorkflow {
	return &cfWorkflow{
		cf:         cfConfig,
		appPath:    appPath,
		org:        org,
		space:      space,
		quota:      quota,
		appName:    appName,
		appCommand: appCommand,
	}
}

func NewWithExistingSpace(cfConfig *config.Cf, appName, appPath, appCommand string) CfWorkflow {
	return &cfWorkflow{
		cf:         cfConfig,
		appPath:    appPath,
		org:        cfConfig.Org,
		space:      cfConfig.Space,
		appName:    appName,
		appCommand: appCommand,
	}
}

func (c *cfWorkflow) Setup(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	commands := []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
	}
	if !c.cf.UseExistingSpace {
		commands = append(commands, ccg.CreateOrg(c.org))
		commands = append(commands, ccg.CreateQuota(c.quota))
		commands = append(commands, ccg.SetQuota(c.org, c.quota))
		commands = append(commands, ccg.CreateSpace(c.org, c.space))
 	}
	return commands
}

func (c *cfWorkflow) Push(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	appInstancesToPush := 2
	if c.cf.UseSingleAppInstance {
		appInstancesToPush = 1
	}

	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
		ccg.Target(c.org, c.space),
		ccg.Push(c.appName, c.cf.AppDomain, c.appPath, c.appCommand, appInstancesToPush),
	}
}

func (c *cfWorkflow) Delete(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
		ccg.Target(c.org, c.space),
		ccg.Delete(c.appName),
	}
}

func (c *cfWorkflow) TearDown(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	commands := []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
	}
	
	if !c.cf.UseExistingSpace {
		commands = append(commands, ccg.DeleteOrg(c.org))
		commands = append(commands, ccg.DeleteQuota(c.quota))
	}
	
	return append(commands, ccg.LogOut())
}

func (c *cfWorkflow) RecentLogs(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
		ccg.Target(c.org, c.space),
		ccg.RecentLogs(c.appName),
	}
}

func (c *cfWorkflow) StreamLogs(ctx context.Context, ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
		ccg.Target(c.org, c.space),
		ccg.StreamLogs(ctx, c.appName),
	}
}

func (c *cfWorkflow) MapRoute(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
		ccg.Target(c.org, c.space),
		ccg.MapRoute(c.appName, c.cf.TCPDomain, c.cf.AvailablePort),
	}
}

func (c *cfWorkflow) CreateAndBindSyslogDrainService(ccg cfCmdGenerator.CfCmdGenerator, serviceName string) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.User, c.cf.Password),
		ccg.Target(c.org, c.space),
		ccg.CreateUserProvidedService(serviceName, fmt.Sprintf("syslog://%s:%d", c.cf.TCPDomain, c.cf.AvailablePort)),
		ccg.BindService(c.appName, serviceName),
		ccg.Restage(c.appName),
	}
}
