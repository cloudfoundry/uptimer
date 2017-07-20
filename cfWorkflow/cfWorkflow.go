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
	cf             *config.CfConfig
	cfCmdGenerator cfCmdGenerator.CfCmdGenerator

	appUrl  string
	appPath string
	org     string
	space   string
	quota   string
	appName string
}

func (c *cfWorkflow) AppUrl() string {
	return c.appUrl
}

func New(cfConfig *config.CfConfig, cfCmdGenerator cfCmdGenerator.CfCmdGenerator, org, space, quota, appName, appPath string) CfWorkflow {
	appUrl := fmt.Sprintf("https://%s.%s", appName, cfConfig.AppDomain)

	return &cfWorkflow{
		cf:             cfConfig,
		cfCmdGenerator: cfCmdGenerator,
		appUrl:         appUrl,
		appPath:        appPath,
		org:            org,
		space:          space,
		quota:          quota,
		appName:        appName,
	}
}

func (c *cfWorkflow) Setup() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.cfCmdGenerator.Api(c.cf.API),
		c.cfCmdGenerator.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		c.cfCmdGenerator.CreateOrg(c.org),
		c.cfCmdGenerator.CreateSpace(c.org, c.space),
		c.cfCmdGenerator.CreateQuota(c.quota),
		c.cfCmdGenerator.SetQuota(c.org, c.quota),
	}
}

func (c *cfWorkflow) Push() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.cfCmdGenerator.Api(c.cf.API),
		c.cfCmdGenerator.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		c.cfCmdGenerator.Target(c.org, c.space),
		c.cfCmdGenerator.Push(c.appName, c.appPath),
	}
}

func (c *cfWorkflow) Delete() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.cfCmdGenerator.Api(c.cf.API),
		c.cfCmdGenerator.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		c.cfCmdGenerator.Target(c.org, c.space),
		c.cfCmdGenerator.Delete(c.appName),
	}
}

func (c *cfWorkflow) TearDown() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.cfCmdGenerator.Api(c.cf.API),
		c.cfCmdGenerator.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		c.cfCmdGenerator.DeleteOrg(c.org),
		c.cfCmdGenerator.DeleteQuota(c.quota),
		c.cfCmdGenerator.LogOut(),
	}
}

func (c *cfWorkflow) RecentLogs() []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.cfCmdGenerator.Api(c.cf.API),
		c.cfCmdGenerator.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		c.cfCmdGenerator.Target(c.org, c.space),
		c.cfCmdGenerator.RecentLogs(c.appName),
	}
}

func (c *cfWorkflow) StreamLogs(ctx context.Context) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		c.cfCmdGenerator.Api(c.cf.API),
		c.cfCmdGenerator.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		c.cfCmdGenerator.Target(c.org, c.space),
		c.cfCmdGenerator.StreamLogs(ctx, c.appName),
	}
}
