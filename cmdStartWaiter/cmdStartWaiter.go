package cmdStartWaiter

import (
	"io"
)

//go:generate counterfeiter . CmdStartWaiter

// CmdStartWaiter is a subset of the interface satisfied by exec.Cmd
type CmdStartWaiter interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}
