package cmdRunner_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	. "github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter/cmdStartWaiterfakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CmdRunner", func() {
	var (
		outBuf *bytes.Buffer
		errBuf *bytes.Buffer

		fakeCmdStartWaiter *cmdStartWaiterfakes.FakeCmdStartWaiter

		runner CmdRunner
	)

	BeforeEach(func() {
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		fakeCmdStartWaiter = &cmdStartWaiterfakes.FakeCmdStartWaiter{}
		fakeCmdStartWaiter.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("")), nil)
		fakeCmdStartWaiter.StderrPipeReturns(io.NopCloser(bytes.NewBufferString("")), nil)

		runner = New(outBuf, errBuf, io.Copy)
	})

	Describe("Run", func() {
		It("starts a command and waits for it to complete", func() {
			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmdStartWaiter.StartCallCount()).To(Equal(1))
			Expect(fakeCmdStartWaiter.WaitCallCount()).To(Equal(1))
		})

		It("returns an error when calling start", func() {
			fakeCmdStartWaiter.StartReturns(fmt.Errorf("something bad"))

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).To(MatchError("something bad"))
		})

		It("returns an error when calling wait", func() {
			fakeCmdStartWaiter.WaitReturns(fmt.Errorf("something bad"))

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).To(MatchError("something bad"))
		})

		It("writes the command's stdout to outWriter", func() {
			fakeCmdStartWaiter.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("something happened on stdout")), nil)

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(Equal("something happened on stdout"))
		})

		It("returns an error when failing to write the command's stdout to outWriter", func() {
			fakeCmdStartWaiter.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("")), fmt.Errorf("something bad happened"))

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).To(MatchError("something bad happened"))
		})

		It("writes the command's stderr to errWriter", func() {
			fakeCmdStartWaiter.StderrPipeReturns(io.NopCloser(bytes.NewBufferString("something happened on stderr")), nil)

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).NotTo(HaveOccurred())
			Expect(errBuf.String()).To(Equal("something happened on stderr"))
		})

		It("returns an error when failing to write the command's stderr to errWriter", func() {
			fakeCmdStartWaiter.StderrPipeReturns(io.NopCloser(bytes.NewBufferString("")), fmt.Errorf("something bad happened"))

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).To(MatchError("something bad happened"))
		})

		It("returns error separately if there was an io.Copy error on stdout", func() {
			mockCopy := func(io.Writer, io.Reader) (int64, error) {
				return 0, fmt.Errorf("i failed on first copyfunc")
			}
			runner = New(outBuf, errBuf, mockCopy)

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).To(MatchError("i failed on first copyfunc"))
		})

		It("returns error separately if there was an io.Copy error on stdout", func() {
			timesCalled := 0
			mockCopy := func(io.Writer, io.Reader) (int64, error) {
				if timesCalled == 1 {
					return 0, fmt.Errorf("i failed on second copyfunc")
				}
				timesCalled++

				return 0, nil
			}
			runner = New(outBuf, errBuf, mockCopy)

			err := runner.Run(fakeCmdStartWaiter)

			Expect(err).To(MatchError("i failed on second copyfunc"))
		})
	})

	Describe("RunWithContext", func() {
		It("doesn't return an error when the context is canceled", func() {
			ctx, cancelFunc := context.WithCancel(context.Background())
			cancelFunc()
			fakeCmdStartWaiter.WaitReturns(context.Canceled)

			err := runner.RunWithContext(ctx, fakeCmdStartWaiter)

			Expect(err).NotTo(HaveOccurred())
		})

		It("doesn't return an error when the context times out", func() {
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Nanosecond)
			defer cancelFunc()
			time.Sleep(time.Millisecond)
			fakeCmdStartWaiter.WaitReturns(context.DeadlineExceeded)

			err := runner.RunWithContext(ctx, fakeCmdStartWaiter)

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error if the context was not canceled or timed out", func() {
			ctx := context.Background()
			fakeCmdStartWaiter.WaitReturns(fmt.Errorf("some error dude"))

			err := runner.RunWithContext(ctx, fakeCmdStartWaiter)

			Expect(err).To(MatchError("some error dude"))
		})
	})

	Describe("RunInSequence", func() {
		var (
			fakeCmdStartWaiter2 *cmdStartWaiterfakes.FakeCmdStartWaiter
		)

		BeforeEach(func() {
			fakeCmdStartWaiter2 = &cmdStartWaiterfakes.FakeCmdStartWaiter{}
			fakeCmdStartWaiter2.StderrPipeReturns(io.NopCloser(bytes.NewBufferString("")), nil)

			fakeCmdStartWaiter.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("1")), nil)
			fakeCmdStartWaiter2.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("2")), nil)
		})

		It("runs commands in sequence", func() {
			err := runner.RunInSequence(fakeCmdStartWaiter, fakeCmdStartWaiter2)

			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(Equal("12"))
		})

		It("returns the error produced by the first command", func() {
			fakeCmdStartWaiter.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("")), fmt.Errorf("something bad happened"))
			fakeCmdStartWaiter2.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("")), fmt.Errorf("something even worse happened"))

			err := runner.RunInSequence(fakeCmdStartWaiter, fakeCmdStartWaiter2)

			Expect(err).To(MatchError("something bad happened"))
			Expect(outBuf.String()).To(BeEmpty())
		})

		It("runs until it encounters an error, returning that error", func() {
			fakeCmdStartWaiter2.StdoutPipeReturns(io.NopCloser(bytes.NewBufferString("")), fmt.Errorf("something even worse happened"))

			err := runner.RunInSequence(fakeCmdStartWaiter, fakeCmdStartWaiter2)

			Expect(err).To(MatchError("something even worse happened"))
			Expect(outBuf.String()).To(Equal("1"))
		})
	})
})
