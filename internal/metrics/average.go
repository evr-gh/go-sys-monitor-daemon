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

func averageDisksLoad(stats []*models.DisksLoad) *models.DisksLoad {
	if len(stats) == 0 {
		return nil
	}

	diskSums := make(map[string]*struct {
		tpsSum      float64
		kpsReadSum  float64
		kpsWriteSum float64
		count       int
	})

	for _, stat := range stats {
		for _, disk := range stat.DisksLoad {
			if _, ok := diskSums[disk.Name]; !ok {
				diskSums[disk.Name] = &struct {
					tpsSum      float64
					kpsReadSum  float64
					kpsWriteSum float64
					count       int
				}{}
			}
			diskSums[disk.Name].tpsSum += disk.Tps
			diskSums[disk.Name].kpsReadSum += disk.KpsRead
			diskSums[disk.Name].kpsWriteSum += disk.KpsWrite
			diskSums[disk.Name].count++
		}
	}

	result := make([]models.DiskLoad, 0, len(diskSums))
	for fsName, sums := range diskSums {
		result = append(result, models.DiskLoad{
			Name:     fsName,
			Tps:      round(sums.tpsSum / float64(sums.count)),
			KpsRead:  round(sums.kpsReadSum / float64(sums.count)),
			KpsWrite: round(sums.kpsWriteSum / float64(sums.count)),
		})
	}

	return &models.DisksLoad{DisksLoad: result}
}

func averageDisksStats(stats []*models.DiskStats) *models.DiskStats {
	if len(stats) == 0 {
		return nil
	}

	diskSums := make(map[string]*struct {
		MBUsagePct uint32
		InUsagePct uint32
		count      uint32
	})

	for _, stat := range stats {
		for _, disk := range stat.DiskStats {
			if _, ok := diskSums[disk.Name]; !ok {
				diskSums[disk.Name] = &struct {
					MBUsagePct uint32
					InUsagePct uint32
					count      uint32
				}{}
			}
			diskSums[disk.Name].InUsagePct += disk.InUsagePct
			diskSums[disk.Name].MBUsagePct += disk.MBUsagePct
			diskSums[disk.Name].count++
		}
	}

	result := make([]models.DiskStat, 0, len(diskSums))
	for name, sums := range diskSums {
		result = append(result, models.DiskStat{
			Name:       name,
			InUsagePct: (sums.InUsagePct / sums.count),
			MBUsagePct: (sums.MBUsagePct / sums.count),
		})
	}

	return &models.DiskStats{DiskStats: result}
}
