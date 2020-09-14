package config

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/jessevdk/go-flags"
)

// Config is a structure that holds the loaded configuration file
type Config struct {
	AWS	credentials `koanf:"aws"`
	Server	app         `koanf:"server"`
}

type credentials struct {
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
	Region    string `koanf:"region"`
}

type app struct {
	Address string `koanf:"address"`
}

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Path to the configuration file" default:"config.toml"`
}

var opt options
var parser = flags.NewParser(&opt, flags.Default)

// GetConfig reads and returns the configuration file
func GetConfig() Config {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
	var cfg = Config{}
	var koanf = koanf.New(".")
	if err := koanf.Load(file.Provider(opt.ConfigFile), toml.Parser()); err != nil {
		log.Fatalf("Error loading configuration file: %v", err)
	}
	if err := koanf.Unmarshal("", &cfg); err != nil {
		log.Fatalf("Error loading configuration file: %v", err)
	}

	return cfg
}
