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
	TCPDomain() string
	TCPPort() int

	Setup(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	Push(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	PushNoRoute(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	Delete(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	TearDown(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	RecentLogs(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	AppStats(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	StreamLogs(context.Context, cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter

	MapSyslogRoute(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	MapTCPRoute(cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter
	CreateAndBindSyslogDrainService(cfCmdGenerator.CfCmdGenerator, string) []cmdStartWaiter.CmdStartWaiter
}

type cfWorkflow struct {
	cf *config.Cf

	appPath string
	org     string
	space   string
	quota   string
	appName string
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

func (c *cfWorkflow) TCPDomain() string {
	return fmt.Sprintf(c.cf.TCPDomain)
}

func (c *cfWorkflow) TCPPort() int {
	return c.cf.TCPPort
}

func New(cfConfig *config.Cf, org, space, quota, appName, appPath string) CfWorkflow {
	return &cfWorkflow{
		cf:      cfConfig,
		appPath: appPath,
		org:     org,
		space:   space,
		quota:   quota,
		appName: appName,
	}
}

func (c *cfWorkflow) Setup(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	ret := []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.CreateOrg(c.org),
	}

	if c.cf.IsolationSegment != "" {
		ret = append(ret,
			ccg.EnableOrgIsolation(c.org, c.cf.IsolationSegment),
			ccg.SetOrgDefaultIsolationSegment(c.org, c.cf.IsolationSegment),
		)
	}
	ret = append(ret, ccg.CreateSpace(c.org, c.space))

	if c.quota != "" {
		ret = append(ret, ccg.CreateQuota(c.quota), ccg.SetQuota(c.org, c.quota))
	}

	return ret
}

func (c *cfWorkflow) Push(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	appInstancesToPush := 2
	if c.cf.UseSingleAppInstance {
		appInstancesToPush = 1
	}

	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.Push(c.appName, c.appPath, appInstancesToPush, false),
	}
}

func (c *cfWorkflow) PushNoRoute(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	appInstancesToPush := 2
	if c.cf.UseSingleAppInstance {
		appInstancesToPush = 1
	}

	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.Push(c.appName, c.appPath, appInstancesToPush, true),
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
	ret := []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.DeleteOrg(c.org),
	}

	if c.quota != "" {
		ret = append(ret, ccg.DeleteQuota(c.quota))
	}
	ret = append(ret, ccg.LogOut())

	return ret
}

func (c *cfWorkflow) AppStats(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.AppStats(c.appName),
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

func (c *cfWorkflow) MapSyslogRoute(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.MapRoute(c.appName, c.cf.TCPDomain, c.cf.AvailablePort),
	}
}

func (c *cfWorkflow) MapTCPRoute(ccg cfCmdGenerator.CfCmdGenerator) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.MapRoute(c.appName, c.cf.TCPDomain, c.cf.TCPPort),
	}
}

func (c *cfWorkflow) CreateAndBindSyslogDrainService(ccg cfCmdGenerator.CfCmdGenerator, serviceName string) []cmdStartWaiter.CmdStartWaiter {
	return []cmdStartWaiter.CmdStartWaiter{
		ccg.Api(c.cf.API),
		ccg.Auth(c.cf.AdminUser, c.cf.AdminPassword),
		ccg.Target(c.org, c.space),
		ccg.CreateUserProvidedService(serviceName, fmt.Sprintf("syslog://%s:%d", c.cf.TCPDomain, c.cf.AvailablePort)),
		ccg.BindService(c.appName, serviceName),
		ccg.Restage(c.appName),
	}
}
