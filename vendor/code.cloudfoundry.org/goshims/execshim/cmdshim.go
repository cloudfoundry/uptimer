package execshim

import (
	"bytes"
	"os/exec"
	"io"
	"syscall"
)

type cmdShim struct {
*exec.Cmd
}

func (c *cmdShim) Start() error {
return c.Cmd.Start()
}

func (c *cmdShim) StdoutPipe() (io.ReadCloser, error) {
return c.Cmd.StdoutPipe()
}

func (c *cmdShim) StderrPipe() (io.ReadCloser, error) {
return c.Cmd.StderrPipe()
}

func (c *cmdShim) Wait() error {
return c.Cmd.Wait()
}

func (c *cmdShim) Run() error {
return c.Cmd.Run()
}

func (c *cmdShim) CombinedOutput() ([]byte, error) {
return c.Cmd.CombinedOutput()
}

func (c *cmdShim) SysProcAttr() *syscall.SysProcAttr {
return c.Cmd.SysProcAttr
}

func (c *cmdShim) Pid() int {
	return c.Cmd.Process.Pid
}

func (c *cmdShim) SetStdout(b *bytes.Buffer) {
	c.Cmd.Stdout = b
}

func (c *cmdShim) SetStderr(b *bytes.Buffer) {
	c.Cmd.Stderr = b
}

func (c *cmdShim) SetEnv(rhs []string) {
	c.Cmd.Env = rhs
}
