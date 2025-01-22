package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database Database
	Name     string `yaml:"name"`
	Telegram Telegram
}

type Telegram struct {
	Token string `yaml:"token"`
	Debug bool   `yaml:"debug"`
}

type Database struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
	Cache   string `yaml:"cache"`
	Schema  string `yaml:"schema"`
	MaxConn int    `yaml:"max_conn"`
}

var config *Config

func NewConfig() (*Config, error) {
	filename := "./config/config.yaml"
	if envFilename := os.Getenv("CONFIG_PATH"); envFilename != "" {
		filename = envFilename
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	config = &conf
	return config, nil
}

func Get() *Config {
	return config
}
