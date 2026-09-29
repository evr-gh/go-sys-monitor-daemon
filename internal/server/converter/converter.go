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
