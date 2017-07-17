package cmdStartWaiter

import (
	"io"
	"os/exec"
)

// CmdStartWaiter is a subset of the interface satisfied by exec.Cmd + Kill
type CmdStartWaiter interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Kill() error
}

type cmdStartWaiter struct {
	*exec.Cmd
}

func (c *cmdStartWaiter) Kill() error {
	return c.Cmd.Process.Kill()
}

func New(cmd *exec.Cmd) CmdStartWaiter {
	return &cmdStartWaiter{
		Cmd: cmd,
	}
}
