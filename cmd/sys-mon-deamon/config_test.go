package main

import (
	"testing"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cmdConfig, err := readConfig("../../configs/go-sys-monitor-daemon.test.yaml")
		require.NoError(t, err)

		require.Equal(t, "1.2.3.4", cmdConfig.RPC.Host)
		require.Equal(t, uint16(5000), cmdConfig.RPC.Port)

		require.Equal(t, interfaces.LogLevel("INFO"), cmdConfig.Logger.Level)

		require.Equal(t, int64(400), cmdConfig.Stats.AveragingPeriodLimit)
		require.Equal(t, true, cmdConfig.Stats.LoadAverage)
		require.Equal(t, false, cmdConfig.Stats.CPUStats)
		require.Equal(t, true, cmdConfig.Stats.DisksLoad)
		require.Equal(t, false, cmdConfig.Stats.DisksStats)
		require.Equal(t, false, cmdConfig.Stats.NetworkConnStats)
		require.Equal(t, true, cmdConfig.Stats.NetworkTopTalkers)
	})

	t.Run("no conf file", func(t *testing.T) {
		cmdConfig, err := readConfig("")
		require.Error(t, err)
		require.Equal(t, "не задан файл конфигурации (--config <Path to configuration file>)", err.Error())
		require.Nil(t, cmdConfig)
	})

	t.Run("no exist conf file", func(t *testing.T) {
		cmdConfig, err := readConfig("../../configs/go-sys-monitor-daemon.not-exist.yaml")
		require.Error(t, err)
		//nolint:lll
		require.Equal(t, "не удалось открыть конфигурационный файл (../../configs/go-sys-monitor-daemon.not-exist.yaml): open ../../configs/go-sys-monitor-daemon.not-exist.yaml: no such file or directory", err.Error())
		require.Nil(t, cmdConfig)
	})
	t.Run("invalid file", func(t *testing.T) {
		cmdConfig, err := readConfig("../../configs/go-sys-monitor-daemon.invalid.yaml")
		require.Error(t, err)
		//nolint:lll
		require.Equal(t, "не удалось разобрать конфигурацию из файла \"../../configs/go-sys-monitor-daemon.invalid.yaml\": decoding failed due to the following error(s):\n\n'rpc.port' cannot parse value as 'uint16': strconv.ParseUint: invalid syntax", err.Error())
		require.Nil(t, cmdConfig)
	})
}
