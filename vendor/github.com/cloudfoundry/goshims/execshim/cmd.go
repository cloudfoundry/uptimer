package execshim

import (
	"bytes"
	"io"
	"syscall"
)

//go:generate counterfeiter -o exec_fake/fake_cmd.go . Cmd

type Cmd interface {
	Start() error
	SetStdout(*bytes.Buffer)
	SetStderr(*bytes.Buffer)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Wait() error
	SetEnv([]string) 
	Run() error
	CombinedOutput() ([]byte, error)
    Pid() int
	SysProcAttr() *syscall.SysProcAttr
}
