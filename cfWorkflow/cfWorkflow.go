package cfWorkflow

import (
	"fmt"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/config"

	"github.com/satori/go.uuid"
)

type CfWorkflow interface {
	Setup() []cmdRunner.CmdStartWaiter
	TearDown() []cmdRunner.CmdStartWaiter
	AppUrl() string
}

type cfWorkflow struct {
	Cf             *config.CfConfig
	CfCmdGenerator cfCmdGenerator.CfCmdGenerator
	appUrl         string
}

func New(cfConfig *config.CfConfig, cfCmdGenerator cfCmdGenerator.CfCmdGenerator) CfWorkflow {
	if cfConfig.Org == "" {
		cfConfig.Org = fmt.Sprintf("uptimer-org-%s", uuid.NewV4().String())
	}

	if cfConfig.Space == "" {
		cfConfig.Space = fmt.Sprintf("uptimer-space-%s", uuid.NewV4().String())
	}

	if cfConfig.AppName == "" {
		cfConfig.AppName = fmt.Sprintf("uptimer-app-%s", uuid.NewV4().String())
	}

	return &cfWorkflow{
		Cf:             cfConfig,
		CfCmdGenerator: cfCmdGenerator,
		appUrl:         fmt.Sprintf("https://%s.%s", cfConfig.AppName, cfConfig.AppDomain),
	}
}

func (c *cfWorkflow) Setup() []cmdRunner.CmdStartWaiter {
	return []cmdRunner.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.CreateOrg(c.Cf.Org),
		c.CfCmdGenerator.CreateSpace(c.Cf.Org, c.Cf.Space),
		c.CfCmdGenerator.Target(c.Cf.Org, c.Cf.Space),
		c.CfCmdGenerator.Push(c.Cf.AppName, c.Cf.AppPath),
	}
}

func (c *cfWorkflow) TearDown() []cmdRunner.CmdStartWaiter {
	return []cmdRunner.CmdStartWaiter{
		c.CfCmdGenerator.Api(c.Cf.API),
		c.CfCmdGenerator.Auth(c.Cf.AdminUser, c.Cf.AdminPassword),
		c.CfCmdGenerator.DeleteOrg(c.Cf.Org),
		c.CfCmdGenerator.LogOut(),
	}
}

func (c *cfWorkflow) AppUrl() string {
	return c.appUrl
}
