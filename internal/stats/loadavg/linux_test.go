//go:build linux

package loadavg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLoadAvg(t *testing.T) {
	t.Run("should parse proc loadavg output correctly", func(t *testing.T) {
		exec := func(_ context.Context, command string, args []string) (string, error) {
			require.Equal(t, "cat", command)
			require.Equal(t, []string{"/proc/loadavg"}, args)
			return `0.32 0.52 0.69 2/1388 17352`, nil
		}

		result, err := GetLoadAvg(context.Background(), exec)

		require.NoError(t, err)
		require.NotNil(t, result)

		require.InDelta(t, 0.32, result.Load1Min, 0.0001)
		require.InDelta(t, 0.52, result.Load5Min, 0.0001)
		require.InDelta(t, 0.69, result.Load15Min, 0.0001)
	})
}

func TestGetStat(t *testing.T) {
	t.Run("test success get stats", func(t *testing.T) {
		load, err := GetStats(context.Background())

		require.NoError(t, err)
		require.NotNil(t, load.Load1Min)
		require.IsType(t, float64(1), load.Load1Min)
		require.NotNil(t, load.Load5Min)
		require.IsType(t, float64(1), load.Load5Min)
		require.NotNil(t, load.Load1Min)
		require.IsType(t, float64(1), load.Load1Min)
	})
}
