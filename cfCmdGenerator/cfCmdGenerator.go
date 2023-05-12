package cfCmdGenerator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

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
	EnableOrgIsolation(org, isolationSegment string) cmdStartWaiter.CmdStartWaiter
	SetOrgDefaultIsolationSegment(org, isolationSegment string) cmdStartWaiter.CmdStartWaiter
	Target(org, space string) cmdStartWaiter.CmdStartWaiter
	Push(name, path string, instances int, noRoute bool) cmdStartWaiter.CmdStartWaiter
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
	cfHome                string
	useBuildpackDetection bool
}

func New(cfHome string, useBuildpackDetection bool) CfCmdGenerator {
	return &cfCmdGenerator{cfHome: cfHome, useBuildpackDetection: useBuildpackDetection}
}

func (c *cfCmdGenerator) setCfHome(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CF_HOME=%s", c.cfHome))
	return cmd
}

func (c *cfCmdGenerator) addCfStagingTimeout(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = append(cmd.Env, "CF_STAGING_TIMEOUT=5")
	return cmd
}

func (c *cfCmdGenerator) Api(url string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "api", url,
			"--skip-ssl-validation",
		),
	)
}

func (c *cfCmdGenerator) Auth(username string, password string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "auth", username, password,
		),
	)
}

func (c *cfCmdGenerator) CreateQuota(quota string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
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
	return c.setCfHome(
		exec.Command(
			"cf", "set-quota", org, quota,
		),
	)
}

func (c *cfCmdGenerator) CreateOrg(org string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "create-org", org,
		),
	)
}

func (c *cfCmdGenerator) EnableOrgIsolation(org, isolationSegment string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "enable-org-isolation", org, isolationSegment,
		),
	)
}
func (c *cfCmdGenerator) SetOrgDefaultIsolationSegment(org, isolationSegment string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "set-org-default-isolation-segment", org, isolationSegment,
		),
	)
}

func (c *cfCmdGenerator) CreateSpace(org string, space string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "create-space", space,
			"-o", org,
		),
	)
}

func (c *cfCmdGenerator) Target(org string, space string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "target",
			"-o", org,
			"-s", space,
		),
	)
}
func (c *cfCmdGenerator) Push(name, path string, instances int, noRoute bool) cmdStartWaiter.CmdStartWaiter {
	args := []string{
		"push", name,
		"-f", "manifest.yml",
		"-i", strconv.Itoa(instances),
	}

	if !c.useBuildpackDetection {
		args = append(args, "-b", "go_buildpack")
	}

	if noRoute {
		args = append(args, "--no-route")
	}

	cmd := exec.Command("cf", args...)
	cmd.Dir = path

	return c.addCfStagingTimeout(
		c.setCfHome(cmd),
	)
}

func (c *cfCmdGenerator) Delete(name string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "delete", name,
			"-f",
			"-r",
		),
	)
}

func (c *cfCmdGenerator) DeleteOrg(org string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "delete-org", org,
			"-f",
		),
	)
}

func (c *cfCmdGenerator) DeleteQuota(quota string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "delete-quota", quota,
			"-f",
		),
	)
}

func (c *cfCmdGenerator) LogOut() cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "logout",
		),
	)
}

func (c *cfCmdGenerator) RecentLogs(appName string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "logs", appName,
			"--recent",
		),
	)
}

func (c *cfCmdGenerator) StreamLogs(ctx context.Context, appName string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.CommandContext(ctx,
			"cf", "logs", appName,
		),
	)
}

func (c *cfCmdGenerator) MapRoute(name, domain string, port int) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "map-route", name, domain,
			"--port", fmt.Sprintf("%d", port),
		),
	)
}

func (c *cfCmdGenerator) CreateUserProvidedService(serviceName, syslogURL string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "create-user-provided-service", serviceName,
			"-l", syslogURL,
		),
	)
}

func (c *cfCmdGenerator) BindService(appName, serviceName string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "bind-service", appName, serviceName,
		),
	)
}

func (c *cfCmdGenerator) Restage(appName string) cmdStartWaiter.CmdStartWaiter {
	return c.setCfHome(
		exec.Command(
			"cf", "restage", appName,
		),
	)
}
