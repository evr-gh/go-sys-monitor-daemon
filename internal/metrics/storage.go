package metrics

import (
	"sync"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/storage"
	memorystorage "github.com/evr-gh/go-sys-monitor-deamon/internal/storage/memory"
)

type Storage struct {
	logger            interfaces.Logger
	mu                sync.RWMutex
	loadAvg           storage.Storage
	cpuStats          storage.Storage
	disksLoad         storage.Storage
	disksStats        storage.Storage
	networkTopTalkers storage.Storage
	networkConnStats  storage.Storage
}

func New(logger interfaces.Logger, sizeLimit int64) *Storage {
	return &Storage{
		logger:            logger,
		mu:                sync.RWMutex{},
		loadAvg:           memorystorage.New(logger, "loadAvg", sizeLimit),
		cpuStats:          memorystorage.New(logger, "cpuStats", sizeLimit),
		disksLoad:         memorystorage.New(logger, "disksLoad", sizeLimit),
		disksStats:        memorystorage.New(logger, "disksStats", sizeLimit),
		networkTopTalkers: memorystorage.New(logger, "networkTopTalkers", sizeLimit),
		networkConnStats:  memorystorage.New(logger, "networkConnStats", sizeLimit),
	}
}

func (m *Storage) StoreLoadAverage(stats *models.LoadAverage, timestamp time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadAvg.Push(stats, timestamp)
}

func (m *Storage) StoreCPUStats(stats *models.CPUStat, timestamp time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cpuStats.Push(stats, timestamp)
}

func (m *Storage) StoreDiskLoad(stats *models.DisksLoad, timestamp time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disksLoad.Push(stats, timestamp)
}

func (m *Storage) StoreDiskStats(stats *models.DiskStats, timestamp time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disksStats.Push(stats, timestamp)
}

func getAverageFromStorage[T any](store storage.Storage, period time.Duration) []T {
	now := time.Now()
	start := now.Add(-period)

	var result []T
	for item := range store.GetElementsSince(start) {
		if metric, ok := item.(T); ok {
			result = append(result, metric)
		}
	}
	return result
}

func (m *Storage) GetAverageLoadAverage(period time.Duration) *models.LoadAverage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := getAverageFromStorage[*models.LoadAverage](m.loadAvg, period)
	res := averageLoadAverage(metrics)
	m.logger.Debug("Подсчитатна средняя загрузка системы по %d значениям: %+v", len(metrics), res)
	return res
}

func (m *Storage) GetAverageCPUStats(period time.Duration) *models.CPUStat {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := getAverageFromStorage[*models.CPUStat](m.cpuStats, period)
	res := averageCPUStats(metrics)
	m.logger.Debug("Подсчитатна средняя загрузка ЦПУ по %d значениям: %+v", len(metrics), res)
	return res
}

func (m *Storage) GetAverageDisksLoad(period time.Duration) *models.DisksLoad {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := getAverageFromStorage[*models.DisksLoad](m.disksLoad, period)

	res := averageDisksLoad(metrics)
	m.logger.Debug("Подсчитатна средняя загрузка дисков по %d значениям: %+v", len(metrics), res)
	return res
}

func (m *Storage) GetAverageDisksStats(period time.Duration) *models.DiskStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := getAverageFromStorage[*models.DiskStats](m.disksStats, period)

	res := averageDisksStats(metrics)
	m.logger.Debug("Подсчитатна статистика по дискам по %d значениям: %+v", len(metrics), res)
	return res
}
