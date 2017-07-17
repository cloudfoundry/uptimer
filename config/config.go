package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	While []*CommandConfig `json:"while"`
	CF    *CfConfig        `json:"cf"`
}

type CommandConfig struct {
	Command     string   `json:"command"`
	CommandArgs []string `json:"command_args"`
}

type CfConfig struct {
	API           string `json:"api"`
	AppDomain     string `json:"app_domain"`
	AdminUser     string `json:"admin_user"`
	AdminPassword string `json:"admin_password"`

	Org     string `json:"org"`
	Space   string `json:"space"`
	AppName string `json:"app_name"`
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
