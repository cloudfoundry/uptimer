package measurement_test

import (
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/fakes"
	. "github.com/cloudfoundry/uptimer/measurement"

	"os/exec"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecentLogs", func() {
	var (
		freq                 time.Duration
		mockClock            *clock.Mock
		commands             []cmdRunner.CmdStartWaiter
		fakeCmdGeneratorFunc func() []cmdRunner.CmdStartWaiter
		fakeCommandRunner    *fakes.FakeCmdRunner

		rlm Measurement
	)

	BeforeEach(func() {
		freq = time.Second
		mockClock = clock.NewMock()
		fakeCommandRunner = &fakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdRunner.CmdStartWaiter {
			return commands
		}

		rlm = NewRecentLogs(freq, mockClock, fakeCmdGeneratorFunc, fakeCommandRunner)
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
			commands = []cmdRunner.CmdStartWaiter{
				exec.Command("foo"),
				exec.Command("bar"),
			}
			err := rlm.Start()
			mockClock.Add(freq)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(BeNumerically(">=", 1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdRunner.CmdStartWaiter{
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
