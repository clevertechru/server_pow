package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Server struct {
		Mode  string `yaml:"mode"`
		Proxy struct {
			Target  string `yaml:"target"`
			Timeout string `yaml:"timeout"`
		} `yaml:"proxy"`
		Quotes struct {
			File string `yaml:"file"`
		} `yaml:"quotes"`
	} `yaml:"server"`
}

func LoadConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *ServerConfig) GetProxyTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Server.Proxy.Timeout)
}
