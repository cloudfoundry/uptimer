package cfCmdGenerator

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

//go:generate counterfeiter . CfCmdGenerator
type CfCmdGenerator interface {
	Api(url string) cmdStartWaiter.CmdStartWaiter
	Auth(username, password string) cmdStartWaiter.CmdStartWaiter
	CreateQuota(quota string) cmdStartWaiter.CmdStartWaiter
	SetQuota(org, quota string) cmdStartWaiter.CmdStartWaiter
	CreateOrg(org string) cmdStartWaiter.CmdStartWaiter
	CreateSpace(org, space string) cmdStartWaiter.CmdStartWaiter
	Target(org, space string) cmdStartWaiter.CmdStartWaiter
	Push(name, domain, path, command string) cmdStartWaiter.CmdStartWaiter
	Delete(name string) cmdStartWaiter.CmdStartWaiter
	DeleteOrg(org string) cmdStartWaiter.CmdStartWaiter
	DeleteQuota(quota string) cmdStartWaiter.CmdStartWaiter
	LogOut() cmdStartWaiter.CmdStartWaiter
	RecentLogs(appName string) cmdStartWaiter.CmdStartWaiter
	StreamLogs(ctx context.Context, appName string) cmdStartWaiter.CmdStartWaiter
	MapRoute(appName, domain string, port int) cmdStartWaiter.CmdStartWaiter
	CreateUserProvidedService(serviceName, syslogURL string) cmdStartWaiter.CmdStartWaiter
	BindService(appName, serviceName string) cmdStartWaiter.CmdStartWaiter
	Restage(appName string) cmdStartWaiter.CmdStartWaiter
}

type cfCmdGenerator struct {
	cfHome string
}

func New(cfHome string) CfCmdGenerator {
	return &cfCmdGenerator{cfHome: cfHome}
}

func (c *cfCmdGenerator) addCfHome(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = append(cmd.Env, fmt.Sprintf("CF_HOME=%s", c.cfHome))
	return cmd
}

func (c *cfCmdGenerator) addCfStagingTimeout(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = append(cmd.Env, "CF_STAGING_TIMEOUT=1")
	return cmd
}

func (c *cfCmdGenerator) Api(url string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "api", url,
			"--skip-ssl-validation",
		),
	)
}

func (c *cfCmdGenerator) Auth(username string, password string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "auth", username, password,
		),
	)
}

func (c *cfCmdGenerator) CreateQuota(quota string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "create-quota", quota,
			"-m", "10G",
			"-i", "1G",
			"-r", "1000",
			"-s", "100",
			"--reserved-route-ports", "1",
		),
	)
}

func (c *cfCmdGenerator) SetQuota(org, quota string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "set-quota", org, quota,
		),
	)
}

func (c *cfCmdGenerator) CreateOrg(org string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "create-org", org,
		),
	)
}

func (c *cfCmdGenerator) CreateSpace(org string, space string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "create-space", space,
			"-o", org,
		),
	)
}

func (c *cfCmdGenerator) Target(org string, space string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "target",
			"-o", org,
			"-s", space,
		),
	)
}

func (c *cfCmdGenerator) Push(name, domain, path, command string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfStagingTimeout(
		c.addCfHome(
			exec.Command(
				"cf", "push", name,
				"-d", domain,
				"-p", path,
				"-b", "binary_buildpack",
				"-c", command,
				"-i", "2",
			),
		),
	)
}

func (c *cfCmdGenerator) Delete(name string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "delete", name,
			"-f",
			"-r",
		),
	)
}

func (c *cfCmdGenerator) DeleteOrg(org string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "delete-org", org,
			"-f",
		),
	)
}

func (c *cfCmdGenerator) DeleteQuota(quota string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "delete-quota", quota,
			"-f",
		),
	)
}

func (c *cfCmdGenerator) LogOut() cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "logout",
		),
	)
}

func (c *cfCmdGenerator) RecentLogs(appName string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "logs", appName,
			"--recent",
		),
	)
}

func (c *cfCmdGenerator) StreamLogs(ctx context.Context, appName string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.CommandContext(ctx,
			"cf", "logs", appName,
		),
	)
}

func (c *cfCmdGenerator) MapRoute(name, domain string, port int) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "map-route", name, domain,
			"--port", fmt.Sprintf("%d", port),
		),
	)
}

func (c *cfCmdGenerator) CreateUserProvidedService(serviceName, syslogURL string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "create-user-provided-service", serviceName,
			"-l", syslogURL,
		),
	)
}

func (c *cfCmdGenerator) BindService(appName, serviceName string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "bind-service", appName, serviceName,
		),
	)
}

func (c *cfCmdGenerator) Restage(appName string) cmdStartWaiter.CmdStartWaiter {
	return c.addCfHome(
		exec.Command(
			"cf", "restage", appName,
		),
	)
}
