package cfCmdGenerator

import (
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdRunner"
)

type CfCmdGenerator interface {
	Api(url string) cmdRunner.CmdStartWaiter
	Auth(username, password string) cmdRunner.CmdStartWaiter
	CreateOrg(org string) cmdRunner.CmdStartWaiter
	CreateSpace(org, space string) cmdRunner.CmdStartWaiter
	Target(org, space string) cmdRunner.CmdStartWaiter
	Push(name, path string) cmdRunner.CmdStartWaiter
	DeleteOrg(org string) cmdRunner.CmdStartWaiter
	LogOut() cmdRunner.CmdStartWaiter
	RecentLogs(appName string) cmdRunner.CmdStartWaiter
}

type cfCmdGenerator struct{}

func New() CfCmdGenerator {
	return &cfCmdGenerator{}
}

func (c *cfCmdGenerator) Api(url string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "api", url, "--skip-ssl-validation")
}

func (c *cfCmdGenerator) Auth(username string, password string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "auth", username, password)
}

func (c *cfCmdGenerator) CreateOrg(org string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "create-org", org)
}

func (c *cfCmdGenerator) CreateSpace(org string, space string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "create-space", space, "-o", org)
}

func (c *cfCmdGenerator) Target(org string, space string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "target", "-o", org, "-s", space)
}

func (c *cfCmdGenerator) Push(name string, path string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "push", name, "-p", path)
}

func (c *cfCmdGenerator) DeleteOrg(org string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "delete-org", org, "-f")
}

func (c *cfCmdGenerator) LogOut() cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "logout")
}

func (c *cfCmdGenerator) RecentLogs(appName string) cmdRunner.CmdStartWaiter {
	return exec.Command("cf", "logs", appName, "--recent")
}
