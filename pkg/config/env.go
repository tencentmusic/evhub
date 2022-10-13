package config

import (
	"flag"
	"os"
	"testing"
)

var (
	// ConfigPath is the path of the application config
	ConfigPath string
	// ConfigName is the name of the application config
	ConfigName string
	// ConfigFileType is the type of the application config
	ConfigFileType string
)

// initialize
func init() {
	flag.StringVar(&ConfigPath, "configPath", "conf/", "config path")
	flag.StringVar(&ConfigName, "configName", "config.toml", "config name")
	flag.StringVar(&ConfigFileType, "configFileType", "toml", "config file type")

	testing.Init()
	flag.Parse()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		ConfigPath = configPath
	}
	configName := os.Getenv("CONFIG_NAME")
	if configName != "" {
		ConfigName = configName
	}
	configFileType := os.Getenv("CONFIG_NAME")
	if configFileType != "" {
		ConfigFileType = configFileType
	}
}
