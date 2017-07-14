package cfCmdGenerator

import (
	"fmt"
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
	Delete(name string) cmdRunner.CmdStartWaiter
	DeleteOrg(org string) cmdRunner.CmdStartWaiter
	LogOut() cmdRunner.CmdStartWaiter
	RecentLogs(appName string) cmdRunner.CmdStartWaiter
}

type cfCmdGenerator struct {
	cfHome string
}

func New(cfHome string) CfCmdGenerator {
	return &cfCmdGenerator{cfHome: cfHome}
}

func (c *cfCmdGenerator) addCfHome(cmd *exec.Cmd) cmdRunner.CmdStartWaiter {
	cmd.Env = []string{fmt.Sprintf("CF_HOME=%s", c.cfHome)}
	return cmd
}

func (c *cfCmdGenerator) Api(url string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "api", url, "--skip-ssl-validation"))
}

func (c *cfCmdGenerator) Auth(username string, password string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "auth", username, password))
}

func (c *cfCmdGenerator) CreateOrg(org string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "create-org", org))
}

func (c *cfCmdGenerator) CreateSpace(org string, space string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "create-space", space, "-o", org))
}

func (c *cfCmdGenerator) Target(org string, space string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "target", "-o", org, "-s", space))
}

func (c *cfCmdGenerator) Push(name string, path string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command(
		"cf", "push", name,
		"-p", path,
		"-b", "binary_buildpack",
		"-c", "./app"))
}

func (c *cfCmdGenerator) Delete(name string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "delete", name, "-f", "-r"))
}

func (c *cfCmdGenerator) DeleteOrg(org string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "delete-org", org, "-f"))
}

func (c *cfCmdGenerator) LogOut() cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "logout"))
}

func (c *cfCmdGenerator) RecentLogs(appName string) cmdRunner.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "logs", appName, "--recent"))
}
