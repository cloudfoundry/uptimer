package measurement_test

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/fakes"
	. "github.com/cloudfoundry/uptimer/measurement"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pushability", func() {
	var (
		freq                 time.Duration
		mockClock            *clock.Mock
		commands             []cmdStartWaiter.CmdStartWaiter
		fakeCmdGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
		fakeCommandRunner    *fakes.FakeCmdRunner

		pm Measurement
	)

	BeforeEach(func() {
		freq = time.Second
		mockClock = clock.NewMock()
		fakeCommandRunner = &fakes.FakeCmdRunner{}
		fakeCmdGeneratorFunc = func() []cmdStartWaiter.CmdStartWaiter {
			return commands
		}

		pm = NewPushability(freq, mockClock, fakeCmdGeneratorFunc, fakeCommandRunner)
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(pm.Name()).To(Equal("App pushability"))
		})
	})

	Describe("Start", func() {
		AfterEach(func() {
			pm.Stop()
		})

		It("runs the generated app push and delete", func() {
			commands = []cmdStartWaiter.CmdStartWaiter{
				cmdStartWaiter.New(exec.Command("foo")),
				cmdStartWaiter.New(exec.Command("bar")),
			}
			err := pm.Start()
			mockClock.Add(freq)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(BeNumerically(">=", 1))
			Expect(fakeCommandRunner.RunInSequenceArgsForCall(0)).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					cmdStartWaiter.New(exec.Command("foo")),
					cmdStartWaiter.New(exec.Command("bar")),
				},
			))
		})

		It("runs the app push and delete commands with given frequency", func() {
			pm.Start()
			mockClock.Add(3 * freq)

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(4))
		})

		It("records the commands that run without an error as success", func() {
			pm.Start()
			mockClock.Add(3 * freq)

			rs, _ := pm.Results()
			Expect(rs.Successful()).To(Equal(4))
		})

		It("records the commands that run with error as failed", func() {
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))

			pm.Start()
			mockClock.Add(3 * freq)

			rs, _ := pm.Results()
			Expect(rs.Failed()).To(Equal(4))
		})

		It("records all of the results", func() {
			pm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(3 * freq)

			rs, _ := pm.Results()
			Expect(rs.Successful()).To(Equal(4))
			Expect(rs.Failed()).To(Equal(3))
			Expect(rs.Total()).To(Equal(7))
		})
	})

	Describe("Stop", func() {
		It("stops the measurement", func() {
			pm.Start()
			mockClock.Add(3 * freq)
			pm.Stop()
			mockClock.Add(3 * freq)

			Expect(fakeCommandRunner.RunInSequenceCallCount()).To(Equal(4))
		})
	})

	Describe("Failed", func() {
		It("returns false when the measurement has succeeded", func() {
			pm.Start()
			mockClock.Add(3 * freq)

			Expect(pm.Failed()).To(BeFalse())
		})

		It("returns true when the measurement has failed", func() {
			pm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(freq)

			Expect(pm.Failed()).To(BeTrue())
		})
	})

	Describe("Summary", func() {
		It("returns a success summary if none failed", func() {
			pm.Start()
			mockClock.Add(3 * freq)
			pm.Stop()

			Expect(pm.Summary()).To(Equal(fmt.Sprintf("SUCCESS(%s): All %d attempts to push and delete an app succeeded", pm.Name(), 4)))
		})

		It("returns a failed summary if there are failures", func() {
			pm.Start()
			mockClock.Add(3 * freq)
			fakeCommandRunner.RunInSequenceReturns(fmt.Errorf("errrrrrooooorrrr"))
			mockClock.Add(3 * freq)
			pm.Stop()

			Expect(pm.Summary()).To(Equal(fmt.Sprintf("FAILED(%s): %d of %d attempts to push and delete an app failed", pm.Name(), 3, 7)))
		})
	})
})
