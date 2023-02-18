package config_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/uptimer/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		configFile *os.File
		err        error
	)

	BeforeEach(func() {
		configFile, err = ioutil.TempFile("", "config")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err = os.Remove(configFile.Name())
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("#Validate", func() {
		It("Returns no error if run_tcp_availability is set to true and tcp_domain and tcp_port are provided", func() {
			cfg := config.Config{
				CF: &config.Cf{
					TCPDomain: "tcp.my-cf.com",
					TCPPort:   1025,
				},
				OptionalTests: config.OptionalTests{RunTcpAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Returns error if run_tcp_availability is set to true, but tcp_domain and available_port are not provided", func() {
			cfg := config.Config{
				CF:            &config.Cf{},
				OptionalTests: config.OptionalTests{RunTcpAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("`cf.tcp_domain` and `cf.tcp_port` must be set in order to run TCP Availability tests."))
		})

		It("Returns error if run_tcp_availability is set to true, but tcp_port is not provided", func() {
			cfg := config.Config{
				CF: &config.Cf{
					TCPDomain: "tcp.my-cf.com",
				},
				OptionalTests: config.OptionalTests{RunTcpAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("`cf.tcp_domain` and `cf.tcp_port` must be set in order to run TCP Availability tests."))
		})

		It("Returns error if run_app_syslog_availability is set to true, but tcp_domain is not provided", func() {
			cfg := config.Config{
				CF: &config.Cf{
					AvailablePort: 1025,
				},
				OptionalTests: config.OptionalTests{RunAppSyslogAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests."))
		})
		It("Returns no error if run_app_syslog_availability is set to true and tcp_domain and available_port are not provided", func() {
			cfg := config.Config{
				CF: &config.Cf{
					TCPDomain:     "tcp.my-cf.com",
					AvailablePort: 1025,
				},
				OptionalTests: config.OptionalTests{RunAppSyslogAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Returns error if run_app_syslog_availability is set to true, but tcp_domain and available_port are not provided", func() {
			cfg := config.Config{
				CF:            &config.Cf{},
				OptionalTests: config.OptionalTests{RunAppSyslogAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests."))
		})

		It("Returns error if run_app_syslog_availability is set to true, but available_port is not provided", func() {
			cfg := config.Config{
				CF: &config.Cf{
					TCPDomain: "tcp.my-cf.com",
				},
				OptionalTests: config.OptionalTests{RunAppSyslogAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests."))
		})

		It("Returns error if run_app_syslog_availability is set to true, but tcp_domain is not provided", func() {
			cfg := config.Config{
				CF: &config.Cf{
					AvailablePort: 1025,
				},
				OptionalTests: config.OptionalTests{RunAppSyslogAvailability: true},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests."))
		})
	})
})
