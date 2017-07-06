package cfCmdGenerator

import (
	"os/exec"
)

type CfCmdGenerator interface {
	Api(url string, skipSslValidation bool) *exec.Cmd
	Auth(username, password string) *exec.Cmd
	CreateOrg(org string) *exec.Cmd
	CreateSpace(org, space string) *exec.Cmd
	Target(org, space string) *exec.Cmd
	Push(name, path string) *exec.Cmd
	DeleteOrg(org string) *exec.Cmd
}

type cfCmdGenerator struct{}

func New() CfCmdGenerator {
	return &cfCmdGenerator{}
}

func (c *cfCmdGenerator) Api(url string, skipSslValidation bool) *exec.Cmd {
	if skipSslValidation {
		return exec.Command("cf", "api", url, "--skip-ssl-validation")
	}

	return exec.Command("cf", "api", url)
}

func (c *cfCmdGenerator) Auth(username string, password string) *exec.Cmd {
	return exec.Command("cf", "auth", username, password)
}

func (c *cfCmdGenerator) CreateOrg(org string) *exec.Cmd {
	return exec.Command("cf", "create-org", org)
}

func (c *cfCmdGenerator) CreateSpace(org string, space string) *exec.Cmd {
	return exec.Command("cf", "create-space", org, space)
}

func (c *cfCmdGenerator) Target(org string, space string) *exec.Cmd {
	return exec.Command("cf", "target", "-o", org, "-s", space)
}

func (c *cfCmdGenerator) Push(name string, path string) *exec.Cmd {
	return exec.Command("cf", "push", name, "-p", path)
}

func (c *cfCmdGenerator) DeleteOrg(org string) *exec.Cmd {
	return exec.Command("cf", "delete-org", org, "-f")
}
