package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	config "github.com/evr-gh/go-sys-monitor-deamon/internal/config"
	"github.com/spf13/viper"
)

var ErrNoConfFile = errors.New("не задан файл конфигурации (--config <Path to configuration file>)")

func readConfig(configFile string) (*config.DaemonConfig, error) {
	if configFile == "" {
		return nil, ErrNoConfFile
	}

	v := viper.New()

	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть конфигурационный файл (%s): %w", configFile, err)
	}
	defer file.Close()

	if err := v.ReadConfig(file); err != nil {
		return nil, fmt.Errorf("не удалось считать конфигурацию из файла %q: %w", configFile, err)
	}

	cfg := config.NewDaemonConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("не удалось разобрать конфигурацию из файла %q: %w", configFile, err)
	}

	return cfg, nil
}
