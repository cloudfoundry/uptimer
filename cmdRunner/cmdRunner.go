package cmdRunner

import (
	"io"
)

// cmdStartWaiter is a subset of the interface satisfied by exec.Cmd
type cmdStartWaiter interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

type CmdRunner interface {
	Run() error
}

type cmdRunner struct {
	CmdStartWaiter cmdStartWaiter
	OutWriter      io.Writer
	ErrWriter      io.Writer
	CopyFunc       copyFunc
}

type copyFunc func(io.Writer, io.Reader) (int64, error)

func New(cmdStartWaiter cmdStartWaiter, outWriter, errWriter io.Writer, copyFunc copyFunc) CmdRunner {
	return &cmdRunner{
		CmdStartWaiter: cmdStartWaiter,
		OutWriter:      outWriter,
		ErrWriter:      errWriter,
		CopyFunc:       copyFunc,
	}
}

func (r *cmdRunner) Run() error {
	stdoutPipe, err := r.CmdStartWaiter.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := r.CmdStartWaiter.StderrPipe()
	if err != nil {
		return err
	}

	if err := r.CmdStartWaiter.Start(); err != nil {
		return err
	}

	if _, err := r.CopyFunc(r.OutWriter, stdoutPipe); err != nil {
		return err
	}

	if _, err := r.CopyFunc(r.ErrWriter, stderrPipe); err != nil {
		return err
	}

	if err := r.CmdStartWaiter.Wait(); err != nil {
		return err
	}

	return nil
}
