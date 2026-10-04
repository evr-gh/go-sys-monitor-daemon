//go:build linux

package cpustats

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDisksLoad(t *testing.T) {
	t.Run("should parse iostat output correctly", func(t *testing.T) {
		exec := func(_ context.Context, command string, args []string) (string, error) {
			require.Equal(t, "iostat", command)
			require.Equal(t, []string{"-c"}, args)
			return `Linux 6.1.166-1-generic (alm01)         29.09.2026      _x86_64_        (12 CPU)

avg-cpu:  %user   %nice %system %iowait  %steal   %idle
           1,51    0,00    0,40    0,04    0,00   98,05`, nil
		}

		result, err := GetCPUStat(context.Background(), exec)

		require.NoError(t, err)
		require.NotNil(t, result)

		require.InDelta(t, 1.51, result.UserMode, 0.0001)
		require.InDelta(t, 0.40, result.SystemMode, 0.0001)
		require.InDelta(t, 98.05, result.Idle, 0.0001)
	})
}

func TestGetStat(t *testing.T) {
	t.Run("test success get stats", func(t *testing.T) {
		cpuStats, err := GetStats(context.Background())
		require.NoError(t, err)
		require.NotNil(t, cpuStats.UserMode)
		require.IsType(t, float64(1), cpuStats.UserMode)
		require.NotNil(t, cpuStats.SystemMode)
		require.IsType(t, float64(1), cpuStats.SystemMode)
		require.NotNil(t, cpuStats.Idle)
		require.IsType(t, float64(1), cpuStats.Idle)
	})
}
