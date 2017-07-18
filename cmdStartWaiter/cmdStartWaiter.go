package cmdStartWaiter

import (
	"io"
)

// CmdStartWaiter is a subset of the interface satisfied by exec.Cmd
//go:generate counterfeiter . CmdStartWaiter
type CmdStartWaiter interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}
