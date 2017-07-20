package measurement_test

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/appLogValidator/appLogValidatorfakes"
	"github.com/cloudfoundry/uptimer/cmdRunner/cmdRunnerfakes"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	. "github.com/cloudfoundry/uptimer/measurement"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecentLogs", func() {
	var (
		freq                 time.Duration
		mockClock            *clock.Mock
		commands             []cmdStartWaiter.CmdStartWaiter
		logger               *log.Logger
		logBuf               *bytes.Buffer
		fakeAppLogValidator  *appLogValidatorfakes.FakeAppLogValidator
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *cmdRunnerfakes.FakeCmdRunner
		outBuf               *bytes.Buffer
		errBuf               *bytes.Buffer

		rlm Measurement
	)

	BeforeEach(func() {
		freq = time.Second
		mockClock = clock.NewMock()
		logBuf = bytes.NewBuffer([]byte{})
		outBuf = bytes.NewBuffer([]byte{})
		errBuf = bytes.NewBuffer([]byte{})

		fakeAppLogValidator = &appLogValidatorfakes.FakeAppLogValidator{}
		fakeAppLogValidator.IsNewerReturns(true, nil)

		fakeCommandRunner = &cmdRunnerfakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}
		logger = log.New(logBuf, "", 0)

		rlm = NewRecentLogs(logger, freq, mockClock, fakeCmdGeneratorFunc, fakeCommandRunner, outBuf, errBuf, fakeAppLogValidator)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(rlm.Name()).To(Equal("Recent logs fetching"))
		})
	})

	Describe("Start", func() {
		AfterEach(func() {
			rlm.Stop()
		})

		It("runs the generated recent logs commands", func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}
			err := rlm.Start()
			mockClock.Add(freq)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(BeNumerically(">=", 1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					exec.Command("foo"),
					exec.Command("bar"),
				},
			))
		})

		It("runs the recent logs commands with given frequency", func() {
			rlm.Start()
			mockClock.Add(3 * freq)

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(4))
		})

		It("records the commands that run without an error as success", func() {
			rlm.Start()
			mockClock.Add(3 * freq)

			rs, _ := rlm.Results()
			Expect(rs.Successful()).To(Equal(4))
		})

		It("records failure when the app logs are not in order", func() {
			fakeAppLogValidator.IsNewerReturns(false, nil)

			rlm.Start()
			mockClock.Add(freq - time.Nanosecond)

			rs, _ := rlm.Results()
			Expect(rs.Failed()).To(Equal(1))
		})

		It("records failure when the app log validator returns an error", func() {
			fakeAppLogValidator.IsNewerReturns(true, fmt.Errorf("oh totally bad news"))

			rlm.Start()
			mockClock.Add(freq - time.Nanosecond)

			rs, _ := rlm.Results()
			Expect(rs.Failed()).To(Equal(1))
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			rlm.Start()
			mockClock.Add(3 * freq)

			rs, _ := rlm.Results()
			Expect(rs.Failed()).To(Equal(4))
		})

		It("records all of the results", func() {
			rlm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(3 * freq)

			rs, _ := rlm.Results()
			Expect(rs.Successful()).To(Equal(4))
			Expect(rs.Failed()).To(Equal(3))
			Expect(rs.Total()).To(Equal(7))
		})

		It("logs both stdout and stderr when there is an error running the command", func() {
			outBuf.WriteString("heyyy guys")
			errBuf.WriteString("whaaats happening?")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			rlm.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Recent logs fetching): errrrrrooooorrrr\x1b[0m\nstdout:\nheyyy guys\nstderr:\nwhaaats happening?\n\n"))
		})

		It("logs both stdout and stderr when the log validator fails", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, nil)

			rlm.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Recent logs fetching): App log fetched was not newer than previous app log fetched\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("logs both stdout and stderr when the log validator returns an error", func() {
			outBuf.WriteString("yo yo")
			errBuf.WriteString("howayah?")
			fakeAppLogValidator.IsNewerReturns(false, fmt.Errorf("we don't need no stinking numbers"))

			rlm.Start()
			mockClock.Add(freq - time.Nanosecond)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Recent logs fetching): App log validation failed with: we don't need no stinking numbers\x1b[0m\nstdout:\nyo yo\nstderr:\nhowayah?\n\n"))
		})

		It("does not accumulate buffers indefinitely", func() {
			outBuf.WriteString("great success")

			rlm.Start()
			mockClock.Add(freq - time.Nanosecond)

			outBuf.WriteString("first failure")
			errBuf.WriteString("that's some standard error")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("e 1"))
			mockClock.Add(freq)

			outBuf.WriteString("second failure")
			errBuf.WriteString("err-body in the club")
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("e 2"))
			mockClock.Add(freq)

			Expect(logBuf.String()).To(Equal("\x1b[31mFAILURE(Recent logs fetching): e 1\x1b[0m\nstdout:\nfirst failure\nstderr:\nthat's some standard error\n\n\x1b[31mFAILURE(Recent logs fetching): e 2\x1b[0m\nstdout:\nsecond failure\nstderr:\nerr-body in the club\n\n"))
		})
	})

	Describe("Stop", func() {
		It("stops the measurement", func() {
			rlm.Start()
			mockClock.Add(3 * freq)
			rlm.Stop()
			mockClock.Add(3 * freq)

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(4))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			rlm.Start()
			mockClock.Add(3 * freq)

			Expect(rlm.Failed()).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			rlm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(freq)

			Expect(rlm.Failed()).To(BeTrue())
		})
	})

	Describe("Summary", func() {
		It("returns a success summary if none failed", func() {
			rlm.Start()
			mockClock.Add(3 * freq)
			rlm.Stop()

			Expect(rlm.Summary()).To(Equal(fmt.Sprintf("SUCCESS(%s): All %d attempts to fetch recent logs succeeded", rlm.Name(), 4)))
		})

		It("returns a failed summary if there are failures", func() {
			rlm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(3 * freq)
			rlm.Stop()

			Expect(rlm.Summary()).To(Equal(fmt.Sprintf("FAILED(%s): %d of %d attempts to fetch recent logs failed", rlm.Name(), 3, 7)))
		})
	})
})
