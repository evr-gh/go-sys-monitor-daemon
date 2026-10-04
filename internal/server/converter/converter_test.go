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

func TestDisksLoadToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := DisksLoadToProto(nil)
		require.Nil(t, result)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &models.DisksLoad{
			DisksLoad: []models.DiskLoad{
				{
					Name:     "/dev/sda1",
					Tps:      100.5,
					KpsRead:  1024.0,
					KpsWrite: 256.0,
				},
				{
					Name:     "/dev/sdb1",
					Tps:      12.4,
					KpsRead:  456.7,
					KpsWrite: 8.9,
				},
			},
		}
		result := DisksLoadToProto(input)
		require.NotNil(t, result)
		require.Len(t, result.DiskLoad, len(input.DisksLoad))

		for i, disk := range input.DisksLoad {
			require.Equal(t, disk.Name, result.DiskLoad[i].Name)
			require.Equal(t, disk.Tps, result.DiskLoad[i].Tps)
			require.Equal(t, disk.KpsRead, result.DiskLoad[i].KpsRead)
			require.Equal(t, disk.KpsWrite, result.DiskLoad[i].KpsWrite)
		}
	})

	t.Run("empty disks list", func(t *testing.T) {
		input := &models.DisksLoad{
			DisksLoad: []models.DiskLoad{},
		}
		result := DisksLoadToProto(input)
		require.NotNil(t, result)
		require.Empty(t, result.DiskLoad)
	})
}

func TestDiskStatsToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := DiskStatsToProto(nil)
		require.Nil(t, result)
	})

	t.Run("valid input", func(t *testing.T) {
		input := &models.DiskStats{
			DiskStats: []models.DiskStat{
				{
					Name:       "/dev/sda1",
					MBUsagePct: 55,
					InUsagePct: 50,
				},
				{
					Name:       "/dev/sdb1",
					MBUsagePct: 15,
					InUsagePct: 30,
				},
			},
		}
		result := DiskStatsToProto(input)
		require.NotNil(t, result)
		require.Len(t, result.DiskStats, len(input.DiskStats))

		for i, disk := range input.DiskStats {
			require.Equal(t, disk.Name, result.DiskStats[i].Name)
			require.Equal(t, disk.MBUsagePct, result.DiskStats[i].MbUsagePct)
			require.Equal(t, disk.InUsagePct, result.DiskStats[i].InUsagePct)
		}
	})

	t.Run("empty disks list", func(t *testing.T) {
		input := &models.DiskStats{
			DiskStats: []models.DiskStat{},
		}
		result := DiskStatsToProto(input)
		require.NotNil(t, result)
		require.Empty(t, result.DiskStats)
	})
}
