package cmdRunner

import (
	"context"
	"io"

	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

//go:generate counterfeiter . CmdRunner
type CmdRunner interface {
	Run(csw cmdStartWaiter.CmdStartWaiter) error
	RunInSequence(csws ...cmdStartWaiter.CmdStartWaiter) error

	RunWithContext(ctx context.Context, csw cmdStartWaiter.CmdStartWaiter) error
	RunInSequenceWithContext(ctx context.Context, csws ...cmdStartWaiter.CmdStartWaiter) error
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

func (r *cmdRunner) Run(cmdStartWaiter cmdStartWaiter.CmdStartWaiter) error {
	return r.RunWithContext(context.TODO(), cmdStartWaiter)
}
func (r *cmdRunner) RunInSequence(cmdStartWaiters ...cmdStartWaiter.CmdStartWaiter) error {
	return r.RunInSequenceWithContext(context.TODO(), cmdStartWaiters...)
}

func (r *cmdRunner) RunInSequenceWithContext(ctx context.Context, csws ...cmdStartWaiter.CmdStartWaiter) error {
	for _, cmd := range csws {
		if err := r.RunWithContext(ctx, cmd); err != nil {
			return err
		}
	}

	return nil
}

func (r *cmdRunner) RunWithContext(ctx context.Context, csw cmdStartWaiter.CmdStartWaiter) error {
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

	// Ignore error due to context cancelation/timeout
	if err = csw.Wait(); err != nil && !wasCanceledOrTimedOut(ctx) {
		return err
	}

	return nil
}

func wasCanceledOrTimedOut(ctx context.Context) bool {
	return ctx != context.TODO() && (ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded)
}
