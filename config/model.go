package config

import (
	"time"
)

// Config is a structure that holds the loaded configuration file
type Config struct {
	AWS    awsCredentials `koanf:"aws"`
	Server server         `koanf:"server"`
}

type awsCredentials struct {
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
	Region    string `koanf:"region"`
	RoleARN   string `koanf:"role_arn"`
}

type server struct {
	Address      string        `koanf:"address"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Path to the configuration file" default:"config.toml"`
}
