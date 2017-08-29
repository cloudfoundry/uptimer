package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	While           []*Command       `json:"while"`
	CF              *Cf              `json:"cf"`
	AllowedFailures *AllowedFailures `json:"allowed_failures"`
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
}

type AllowedFailures struct {
	AppPushability     int `json:"app_pushability"`
	HttpAvailability   int `json:"http_availability"`
	RecentLogsFetching int `json:"recent_logs_fetching"`
	StreamingLogs      int `json:"streaming_logs"`
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
