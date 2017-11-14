package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	While           []*Command      `json:"while"`
	CF              *Cf             `json:"cf"`
	OptionalTests   OptionalTests   `json:"optional_tests"`
	AllowedFailures AllowedFailures `json:"allowed_failures"`
}

type Command struct {
	Command     string   `json:"command"`
	CommandArgs []string `json:"command_args"`
}

type Cf struct {
	API           string `json:"api"`
	AppDomain     string `json:"app_domain"`
	AdminUser     string `json:"admin_user"`
	AdminPassword string `json:"admin_password"`

	TCPDomain     string `json:"tcp_domain"`
	AvailablePort int    `json:"available_port"`
}

type AllowedFailures struct {
	AppPushability        int `json:"app_pushability"`
	HttpAvailability      int `json:"http_availability"`
	RecentLogs            int `json:"recent_logs"`
	StreamingLogs         int `json:"streaming_logs"`
	AppSyslogAvailability int `json:"app_syslog_availability"`
}

type OptionalTests struct {
	RunAppSyslogAvailability bool `json:"run_app_syslog_availability"`
}

func Load(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	newConfig := &Config{}
	err = json.Unmarshal(data, newConfig)
	if err != nil {
		return nil, err
	}

	return newConfig, nil
}
