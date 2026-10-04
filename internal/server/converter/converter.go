package converter

import (
	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc/grpcapi"
)

func LoadAverageToProto(la *models.LoadAverage) *grpcapi.LoadAverage {
	if la == nil {
		return nil
	}
	return &grpcapi.LoadAverage{
		Load1Min:  la.Load1Min,
		Load5Min:  la.Load5Min,
		Load15Min: la.Load15Min,
	}
}

func CPUStatToProto(cs *models.CPUStat) *grpcapi.CPUStats {
	if cs == nil {
		return nil
	}
	return &grpcapi.CPUStats{
		UserMode:   cs.UserMode,
		SystemMode: cs.SystemMode,
		Idle:       cs.Idle,
	}
}

func DisksLoadToProto(dl *models.DisksLoad) *grpcapi.DisksLoad {
	if dl == nil {
		return nil
	}

	disks := make([]*grpcapi.DiskLoad, len(dl.DisksLoad))
	for i, disk := range dl.DisksLoad {
		disks[i] = &grpcapi.DiskLoad{
			Name:     disk.Name,
			Tps:      disk.Tps,
			KpsRead:  disk.KpsRead,
			KpsWrite: disk.KpsWrite,
		}
	}
	return &grpcapi.DisksLoad{
		DiskLoad: disks,
	}
}

func DiskStatsToProto(ds *models.DiskStats) *grpcapi.DisksStats {
	if ds == nil {
		return nil
	}

	diskStats := make([]*grpcapi.DiskStats, len(ds.DiskStats))
	for i, diskStat := range ds.DiskStats {
		diskStats[i] = &grpcapi.DiskStats{
			Name:       diskStat.Name,
			MbUsagePct: diskStat.MBUsagePct,
			InUsagePct: diskStat.InUsagePct,
		}
	}
	return &grpcapi.DisksStats{
		DiskStats: diskStats,
	}
}
