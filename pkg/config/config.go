package config

import (
	"github.com/spf13/viper"
)

// Config is the configuration for the application
type Config struct {
	ConfigPath     string
	ConfigName     string
	ConfigFileType string
}

// New creates a new Config instance
func New() (*Config, error) {
	return &Config{
		ConfigPath:     ConfigPath,
		ConfigName:     ConfigName,
		ConfigFileType: ConfigFileType,
	}, nil
}

// Load loads the configuration.
func (s *Config) Load(c interface{}) error {
	vp := viper.New()
	vp.SetConfigName(s.ConfigName)
	vp.SetConfigType(s.ConfigFileType)
	vp.AddConfigPath(s.ConfigPath)
	err := vp.ReadInConfig()
	if err != nil {
		return err
	}
	err = vp.Unmarshal(c)
	if err != nil {
		return err
	}
	return nil
}
