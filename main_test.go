package main_test

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/cloudfoundry/uptimer/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("uptimer", func() {
	var (
		cfg     *config.Config
		session *Session
	)

	BeforeEach(func() {
		cfg = &config.Config{
			While: []*config.Command{
				{
					Command:     "sleep",
					CommandArgs: []string{"5"},
				},
			},
			CF: &config.Cf{
				API:           "api.my-cf.com",
				AppDomain:     "my-cf.com",
				AdminUser:     "admin",
				AdminPassword: "pass",
			},
			AllowedFailures: config.AllowedFailures{
				AppPushability:   2,
				HttpAvailability: 5,
				RecentLogs:       2,
				StreamingLogs:    2,
			},
		}
	})

	JustBeforeEach(func() {
		tmpDir := GinkgoT().TempDir()
		f, err := os.Create(tmpDir + "/config.json")
		Expect(err).NotTo(HaveOccurred())
		defer f.Close() //nolint:errcheck
		b, err := json.Marshal(cfg)
		Expect(err).NotTo(HaveOccurred())
		_, err = f.Write(b)
		Expect(err).NotTo(HaveOccurred())
		cmd := exec.Command(uptimerPath, "-configFile", f.Name())
		session, err = Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(Exit())
	})

	Context("when configured to test app syslog availability", func() {
		BeforeEach(func() {
			cfg.OptionalTests.RunAppSyslogAvailability = true
			cfg.AllowedFailures.AppSyslogAvailability = 2
		})

		Context("when tcp_domain and available_port are not provided", func() {
			BeforeEach(func() {
				cfg.CF.AvailablePort = 0
				cfg.CF.TCPDomain = ""
			})

			It("exits with a error code of 1", func() {
				Expect(session.ExitCode()).To(Equal(1))
			})

			It("prints an error", func() {
				Expect(session.Out).To(Say("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests"))
			})
		})
	})

	Context("when configured to test TCP availability", func() {
		BeforeEach(func() {
			cfg.OptionalTests.RunTcpAvailability = true
			cfg.AllowedFailures.TCPAvailability = 2
		})

		Context("when tcp_domain and tcp_port are not provided", func() {
			BeforeEach(func() {
				cfg.CF.TCPDomain = ""
				cfg.CF.TCPPort = 0
			})

			It("exits with a error code of 1", func() {
				Expect(session.ExitCode()).To(Equal(1))
			})

			It("prints an error", func() {
				Expect(session.Out).To(Say("`cf.tcp_domain` and `cf.tcp_port` must be set in order to run TCP Availability tests"))
			})
		})
	})
})
