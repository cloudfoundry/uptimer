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

	Setup(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	Push(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	Delete(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	TearDown(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	RecentLogs(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	StreamLogs(context.Context, cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
}

type cfWorkflow struct {
	cf *config.CfConfig

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

func New(cfConfig *config.CfConfig, org, space, quota, appName, appPath string) CfWorkflow {
	appUrl := fmt.Sprintf("https://%s.%s", appName, cfConfig.AppDomain)

	return &cfWorkflow{
		cf:      cfConfig,
		appUrl:  appUrl,
		appPath: appPath,
		org:     org,
		space:   space,
		quota:   quota,
		appName: appName,
	}
}

func (c *cfWorkflow) Setup(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.CreateOrg(c.org),
		ccg.CreateSpace(c.org, c.space),
		ccg.CreateQuota(c.quota),
		ccg.SetQuota(c.org, c.quota),
	}
}

func (c *cfWorkflow) Push(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.Push(c.appName, c.appPath),
	}
}

func (c *cfWorkflow) Delete(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.Delete(c.appName),
	}
}

func (c *cfWorkflow) TearDown(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.DeleteOrg(c.org),
		ccg.DeleteQuota(c.quota),
		ccg.LogOut(),
	}
}

func (c *cfWorkflow) RecentLogs(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.RecentLogs(c.appName),
	}
}

func (c *cfWorkflow) StreamLogs(ctx context.Context, ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.StreamLogs(ctx, c.appName),
	}
}
