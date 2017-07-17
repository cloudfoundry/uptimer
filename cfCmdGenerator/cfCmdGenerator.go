package cfCmdGenerator

import (
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type CfCmdGenerator interface {
	Api(url string) cmdStartWaiter.CmdStartWaiter
	Auth(username, password string) cmdStartWaiter.CmdStartWaiter
	CreateOrg(org string) cmdStartWaiter.CmdStartWaiter
	CreateSpace(org, space string) cmdStartWaiter.CmdStartWaiter
	Target(org, space string) cmdStartWaiter.CmdStartWaiter
	Push(name, path string) cmdStartWaiter.CmdStartWaiter
	Delete(name string) cmdStartWaiter.CmdStartWaiter
	DeleteOrg(org string) cmdStartWaiter.CmdStartWaiter
	LogOut() cmdStartWaiter.CmdStartWaiter
	RecentLogs(appName string) cmdStartWaiter.CmdStartWaiter
}

type cfCmdGenerator struct {
	cfHome string
}

func New(cfHome string) CfCmdGenerator {
	return &cfCmdGenerator{cfHome: cfHome}
}

func (c *cfCmdGenerator) addCfHome(cmd *exec.Cmd) cmdStartWaiter.CmdStartWaiter {
	cmd.Env = []string{fmt.Sprintf("CF_HOME=%s", c.cfHome)}
	return cmdStartWaiter.New(cmd)
}

func (c *cfCmdGenerator) Api(url string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "api", url, "--skip-ssl-validation"))
}

func (c *cfCmdGenerator) Auth(username string, password string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "auth", username, password))
}

func (c *cfCmdGenerator) CreateOrg(org string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "create-org", org))
}

func (c *cfCmdGenerator) CreateSpace(org string, space string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "create-space", space, "-o", org))
}

func (c *cfCmdGenerator) Target(org string, space string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "target", "-o", org, "-s", space))
}

func (c *cfCmdGenerator) Push(name string, path string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command(
		"cf", "push", name,
		"-p", path,
		"-b", "binary_buildpack",
		"-c", "./app"))
}

func (c *cfCmdGenerator) Delete(name string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "delete", name, "-f", "-r"))
}

func (c *cfCmdGenerator) DeleteOrg(org string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "delete-org", org, "-f"))
}

func (c *cfCmdGenerator) LogOut() cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "logout"))
}

func (c *cfCmdGenerator) RecentLogs(appName string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "logs", appName, "--recent"))
}

func (c *cfCmdGenerator) StreamLogs(appName string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(exec.Command("cf", "logs", appName))
}
