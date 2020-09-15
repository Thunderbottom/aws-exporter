package config

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/jessevdk/go-flags"
)

var opt options
var parser = flags.NewParser(&opt, flags.Default)

// GetConfig reads and returns the configuration file
func GetConfig() Config {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		default:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			log.Fatalf("Error loading configuration file: %v", err)
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
