package config

import "github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"

type LoggerConfig struct {
	Level interfaces.LogLevel `mapstructure:"level" json:"level"`
	File  string              `mapstructure:"file" json:"file"`
}
