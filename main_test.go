package main_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	It("Prints error when when run_app_syslog_availability and tcp_domain and available_port are not provided", func() {
		configFile, err := ioutil.TempFile("", "config")
		Expect(err).ToNot(HaveOccurred())

		defer os.Remove(configFile.Name())

		configJson := `{
    "while": [
        {
            "command": "sleep",
            "command_args": ["5"]
        }
    ],
    "cf": {
        "api": "api.my-cf.com",
        "app_domain": "my-cf.com",
        "admin_user": "admin",
        "admin_password": "PASS"
    },
    "optional_tests": {
      "run_app_syslog_availability": true
    },
    "allowed_failures": {
        "app_pushability": 2,
        "http_availability": 5,
        "recent_logs": 2,
        "streaming_logs": 2,
        "app_syslog_availability": 2
    }
}`
		_, err = configFile.WriteString(configJson)
		Expect(err).ToNot(HaveOccurred())

		session := runCommand("-configFile", configFile.Name())

		Expect(string(session.Out.Contents())).To(ContainSubstring("`cf.tcp_domain` and `cf.available_port` must be set in order to run App Syslog Availability tests."))
	})
})
