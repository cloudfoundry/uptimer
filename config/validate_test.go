package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/uptimer/config"
)

var _ = Describe("Validate", func() {
	var (
		cfg config.Config

		err error
	)

	Context("when measuring TCP availability", func() {
		BeforeEach(func() {
			cfg = config.Config{
				CF: &config.Cf{
					TCPDomain: "tcp.my-cf.com",
					TCPPort:   1025,
				},
				OptionalTests: config.OptionalTests{RunTcpAvailability: true},
			}
		})

		JustBeforeEach(func() {
			err = cfg.Validate()
		})

		It("succeeds", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when neither a TCP port or domain are provided", func() {
			BeforeEach(func() {
				cfg.CF.TCPDomain = ""
				cfg.CF.TCPPort = 0
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("`cf.tcp_domain` and `cf.tcp_port` must be set in order to run TCP Availability tests"))
			})
		})

		Context("when a TCP domain is not provided", func() {
			BeforeEach(func() {
				cfg.CF.TCPDomain = ""
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("`cf.tcp_domain` and `cf.tcp_port` must be set in order to run TCP Availability tests"))
			})
		})

		Context("when a TCP port is not provided", func() {
			BeforeEach(func() {
				cfg.CF.TCPPort = 0
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("`cf.tcp_domain` and `cf.tcp_port` must be set in order to run TCP Availability tests"))
			})
		})
	})

	Context("when measuring app syslog availability", func() {
		BeforeEach(func() {
			cfg = config.Config{
				CF: &config.Cf{
					TCPDomain:     "tcp.my-cf.com",
					AvailablePort: 1025,
				},
				OptionalTests: config.OptionalTests{RunAppSyslogAvailability: true},
			}
		})

		JustBeforeEach(func() {
			err = cfg.Validate()
		})

		It("succeeds", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when neither a TCP domain or available port are provided", func() {
			BeforeEach(func() {
				cfg.CF.TCPDomain = ""
				cfg.CF.AvailablePort = 0
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests"))
			})
		})

		Context("when a TCP domain is not provided", func() {
			BeforeEach(func() {
				cfg.CF.TCPDomain = ""
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests"))
			})
		})

		Context("when an available port is not provided", func() {
			BeforeEach(func() {
				cfg.CF.AvailablePort = 0
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests"))
			})
		})
	})
})
