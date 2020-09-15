package config

import (
	"time"
)

// Config is a structure that holds the loaded configuration file
type Config struct {
	Jobs   []Job          `koanf:"jobs"`
	Server server         `koanf:"server"`
}

// Job is a structure that holds configuration for aws accounts
type Job struct {
	Name    string         `koanf:"name"`
	AWS     awsCredentials `koanf:"aws"`
	Filters []filters      `koanf:"filters"`
}

type awsCredentials struct {
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
	Region    string `koanf:"region"`
	RoleARN   string `koanf:"role_arn"`
}

type filters struct {
	Name  string `koanf:"name"`
	Value string `koanf:"value"`
}

type server struct {
	Address      string        `koanf:"address"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Path to the configuration file" default:"config.toml"`
}
