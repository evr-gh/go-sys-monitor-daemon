package converter

import (
	"testing"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/stretchr/testify/require"
)

func TestLoadAverageToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := LoadAverageToProto(nil)
		require.Nil(t, result)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &models.LoadAverage{
			Load1Min:  1.5,
			Load5Min:  2.0,
			Load15Min: 1.8,
		}
		result := LoadAverageToProto(input)
		require.NotNil(t, result)
		require.Equal(t, input.Load1Min, result.Load1Min)
		require.Equal(t, input.Load5Min, result.Load5Min)
		require.Equal(t, input.Load15Min, result.Load15Min)
	})
}

func TestCPUStatToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := CPUStatToProto(nil)
		require.Nil(t, result)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &models.CPUStat{
			UserMode:   12.3,
			SystemMode: 4.5,
			Idle:       67.8,
		}
		result := CPUStatToProto(input)
		require.NotNil(t, result)
		require.Equal(t, input.UserMode, result.UserMode)
		require.Equal(t, input.SystemMode, result.SystemMode)
		require.Equal(t, input.Idle, result.Idle)
	})
}
