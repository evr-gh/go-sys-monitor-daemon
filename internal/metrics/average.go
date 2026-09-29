package metrics

import (
	"math"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
)

func round(x float64) float64 {
	return math.Round(x*100) / 100
}

func averageLoadAverage(stats []*models.LoadAverage) *models.LoadAverage {
	if len(stats) == 0 {
		return nil
	}

	var sum models.LoadAverage
	for _, stat := range stats {
		sum.Load1Min += stat.Load1Min
		sum.Load5Min += stat.Load5Min
		sum.Load15Min += stat.Load15Min
	}

	count := float64(len(stats))
	return &models.LoadAverage{
		Load1Min:  round(sum.Load1Min / count),
		Load5Min:  round(sum.Load5Min / count),
		Load15Min: round(sum.Load15Min / count),
	}
}

func averageCPUStats(stats []*models.CPUStat) *models.CPUStat {
	if len(stats) == 0 {
		return nil
	}

	var sum models.CPUStat
	for _, stat := range stats {
		sum.UserMode += stat.UserMode
		sum.SystemMode += stat.SystemMode
		sum.Idle += stat.Idle
	}

	count := float64(len(stats))
	return &models.CPUStat{
		UserMode:   round(sum.UserMode / count),
		SystemMode: round(sum.SystemMode / count),
		Idle:       round(sum.Idle / count),
	}
}
