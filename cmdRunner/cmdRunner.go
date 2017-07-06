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
	Run(cmdStartWaiter cmdStartWaiter) error
}

type cmdRunner struct {
	OutWriter io.Writer
	ErrWriter io.Writer
	CopyFunc  copyFunc
}

type copyFunc func(io.Writer, io.Reader) (int64, error)

func New(outWriter, errWriter io.Writer, copyFunc copyFunc) CmdRunner {
	return &cmdRunner{
		OutWriter: outWriter,
		ErrWriter: errWriter,
		CopyFunc:  copyFunc,
	}
}

func (r *cmdRunner) Run(csw cmdStartWaiter) error {
	stdoutPipe, err := csw.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := csw.StderrPipe()
	if err != nil {
		return err
	}

	if err := csw.Start(); err != nil {
		return err
	}

	if _, err := r.CopyFunc(r.OutWriter, stdoutPipe); err != nil {
		return err
	}

	if _, err := r.CopyFunc(r.ErrWriter, stderrPipe); err != nil {
		return err
	}

	if err := csw.Wait(); err != nil {
		return err
	}

	return nil
}
