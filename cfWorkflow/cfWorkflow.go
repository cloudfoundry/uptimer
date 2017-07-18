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
	AppUrl() string

	Setup() []cmdStartWaiter.CmdStartWaiter
	Push() []cmdStartWaiter.CmdStartWaiter
	Delete() []cmdStartWaiter.CmdStartWaiter
	TearDown() []cmdStartWaiter.CmdStartWaiter
	RecentLogs() []cmdStartWaiter.CmdStartWaiter
	StreamLogs(ctx context.Context) []cmdStartWaiter.CmdStartWaiter
}

type cfWorkflow struct {
	Cf             *config.CfConfig
	CfCmdGenerator cfCmdGenerator.CfCmdGenerator

	appUrl  string
	appPath string
	org     string
	space   string
	appName string
}

func (c *cfWorkflow) AppUrl() string {
	return c.appUrl
}

func New(cfConfig *config.CfConfig, cfCmdGenerator cfCmdGenerator.CfCmdGenerator, org, space, appName, appPath string) CfWorkflow {
	appUrl := fmt.Sprintf("https://%s.%s", appName, cfConfig.AppDomain)

	return &cfWorkflow{
		Cf:             cfConfig,
		CfCmdGenerator: cfCmdGenerator,
		appUrl:         appUrl,
		appPath:        appPath,
		org:            org,
		space:          space,
		appName:        appName,
	}
}

func (c *cfWorkflow) Setup() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.CreateOrg(c.org),
		c.CfCmdGenerator.CreateSpace(c.org, c.space),
	}
}

func (c *cfWorkflow) Push() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.Target(c.org, c.space),
		c.CfCmdGenerator.Push(c.appName, c.appPath),
	}
}

func (c *cfWorkflow) Delete() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.Target(c.org, c.space),
		c.CfCmdGenerator.Delete(c.appName),
	}
}

func (c *cfWorkflow) TearDown() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.DeleteOrg(c.org),
		c.CfCmdGenerator.LogOut(),
	}
}

func (c *cfWorkflow) RecentLogs() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.Target(c.org, c.space),
		c.CfCmdGenerator.RecentLogs(c.appName),
	}
}

func (c *cfWorkflow) StreamLogs(ctx context.Context) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.Target(c.org, c.space),
		c.CfCmdGenerator.StreamLogs(ctx, c.appName),
	}
}
