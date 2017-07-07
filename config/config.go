package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Command           string   `json:"command"`
	CommandArgs       []string `json:"commandArgs"`
	Api               string   `json:"api"`
	AdminUsername     string   `json:"adminUsername"`
	AdminPassword     string   `json:"adminPassword"`
	SkipSslValidation bool     `json:"skipSslValidation"`
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
