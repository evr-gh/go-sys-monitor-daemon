package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/config"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/metrics"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/server/converter"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc/grpcapi"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/stats/cpustats"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/stats/disksload"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/stats/diskstats"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/stats/loadavg"
)

type MetricsCollector struct {
	logger      interfaces.Logger
	metrics     *metrics.Storage
	mu          sync.RWMutex
	statsConfig config.StatsConfig
	ctx         context.Context
	cancel      context.CancelFunc
	work        map[grpcapi.StatType]int32
}

func NewMetricsCollector(logger interfaces.Logger, conf config.StatsConfig) *MetricsCollector {
	return &MetricsCollector{
		logger:      logger,
		statsConfig: conf,
		metrics:     metrics.New(logger, conf.AveragingPeriodLimit+1),
		mu:          sync.RWMutex{},
		ctx:         context.Background(),
		work: map[grpcapi.StatType]int32{
			grpcapi.StatType_LOAD_AVERAGE:        0,
			grpcapi.StatType_CPU_STATS:           0,
			grpcapi.StatType_DISKS_LOAD:          0,
			grpcapi.StatType_DISKS_STATS:         0,
			grpcapi.StatType_NETWORK_TOP_TALKERS: 0,
			grpcapi.StatType_NETWORK_CONN_STATS:  0,
		},
	}
}

func (c *MetricsCollector) collectLoadAverage(timestamp time.Time) {
	if !c.statsConfig.LoadAverage {
		return
	}
	c.logger.Debug("collectLoadAverage Start")
	if stats, err := loadavg.GetStats(c.ctx); err == nil {
		c.metrics.StoreLoadAverage(stats, timestamp)
	} else {
		c.logger.Error("Ошибка при сборе метрик по средней загрузке системы")
	}
	c.logger.Debug("collectLoadAverage End")
}

func (c *MetricsCollector) collectCPUSats(timestamp time.Time) {
	if !c.statsConfig.LoadAverage {
		return
	}
	c.logger.Debug("collectCpuSats Start")
	if stats, err := cpustats.GetStats(c.ctx); err == nil {
		c.metrics.StoreCPUStats(stats, timestamp)
	} else {
		c.logger.Error("Ошибка при сборе метрик по ЦПУ")
	}
	c.logger.Debug("collectCpuSats End")
}

func (c *MetricsCollector) collectDiskLoad(timestamp time.Time) {
	if !c.statsConfig.LoadAverage {
		return
	}
	c.logger.Debug("collectDiskLoad Start")
	if stats, err := disksload.GetStats(c.ctx); err == nil {
		c.metrics.StoreDiskLoad(stats, timestamp)
	} else {
		c.logger.Error("Ошибка при сборе метрик по загрузке дистков")
	}
	c.logger.Debug("collectDiskLoad End")
}

func (c *MetricsCollector) collectDiskStats(timestamp time.Time) {
	if !c.statsConfig.LoadAverage {
		return
	}
	_ = timestamp
	c.logger.Debug("collectDiskStats Start")
	if stats, err := diskstats.GetStats(c.ctx); err == nil {
		c.metrics.StoreDiskStats(stats, timestamp)
	} else {
		c.logger.Error("Ошибка при сборе метрик по дискам")
	}
	c.logger.Debug("collectDiskStats End")
}

func (c *MetricsCollector) collectNetworkTopTalkers(timestamp time.Time) {
	if !c.statsConfig.LoadAverage {
		return
	}
	_ = timestamp
	c.logger.Debug("collectNetworkTopTalkers Start")
	// if stats, err := loadavg.GetStats(); err == nil {
	//	c.metrics.StoreLoadAverage(stats, timestamp)
	//	}
	c.logger.Debug("collectNetworkTopTalkers End")
}

func (c *MetricsCollector) collectNetworkConnStats(timestamp time.Time) {
	if !c.statsConfig.LoadAverage {
		return
	}
	_ = timestamp
	c.logger.Debug("collectNetworkConnStats Start")
	// if stats, err := loadavg.GetStats(); err == nil {
	//	c.metrics.StoreLoadAverage(stats, timestamp)
	//	}
	c.logger.Debug("collectNetworkConnStats End")
}

func (c *MetricsCollector) collectMetrics(timestamp time.Time) {
	var wg sync.WaitGroup

	for statType, value := range c.work {
		if value > 0 {
			wg.Add(1)
			go func(statType grpcapi.StatType) {
				defer wg.Done()
				switch statType {
				case grpcapi.StatType_LOAD_AVERAGE:
					c.collectLoadAverage(timestamp)
				case grpcapi.StatType_CPU_STATS:
					c.collectCPUSats(timestamp)
				case grpcapi.StatType_DISKS_LOAD:
					c.collectDiskLoad(timestamp)
				case grpcapi.StatType_DISKS_STATS:
					c.collectDiskStats(timestamp)
				case grpcapi.StatType_NETWORK_TOP_TALKERS:
					c.collectNetworkTopTalkers(timestamp)
				case grpcapi.StatType_NETWORK_CONN_STATS:
					c.collectNetworkConnStats(timestamp)
				}
			}(statType)
		}
	}

	wg.Wait()
}

func (c *MetricsCollector) Run(ctx context.Context) error {
	c.logger.Info("Запуск сбора метрик")

	c.mu.Lock()
	c.ctx, c.cancel = context.WithCancel(ctx)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	c.mu.Unlock()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Завершение сбора метрик по завершению работы сервера")
			return nil
		case <-ticker.C:
			currentTime := time.Now()
			c.collectMetrics(currentTime)
		}
	}
}

func (c *MetricsCollector) Stop() {
	c.logger.Info("Останов сбора метрик")
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *MetricsCollector) RegisterRequest(averagingPeriod int64, statTyes []grpcapi.StatType) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, statType := range statTyes {
		switch statType {
		case grpcapi.StatType_LOAD_AVERAGE:
			if !c.statsConfig.LoadAverage {
				return fmt.Errorf("не собирается статистика по средней загрузке системы")
			}
		case grpcapi.StatType_CPU_STATS:
			if !c.statsConfig.LoadAverage {
				return fmt.Errorf("не собирается статистика по ЦПУ")
			}
		case grpcapi.StatType_DISKS_LOAD:
			if !c.statsConfig.DisksLoad {
				return fmt.Errorf("не собирается статистика по средней загрузке дисков")
			}
		case grpcapi.StatType_DISKS_STATS:
			if !c.statsConfig.DisksStats {
				return fmt.Errorf("не собирается статистика по дискам")
			}
		case grpcapi.StatType_NETWORK_TOP_TALKERS:
			if !c.statsConfig.NetworkTopTalkers {
				return fmt.Errorf("не собирается статистика по тop talkers в сети")
			}
		case grpcapi.StatType_NETWORK_CONN_STATS:
			if !c.statsConfig.NetworkConnStats {
				return fmt.Errorf("не собирается статистика по сетевым соединениям")
			}
		}
	}

	if averagingPeriod > c.statsConfig.AveragingPeriodLimit {
		return fmt.Errorf("период усреднения %d в запросе превышает лимит %d",
			averagingPeriod, c.statsConfig.AveragingPeriodLimit)
	}
	for _, statType := range statTyes {
		c.work[statType]++
	}
	return nil
}

func (c *MetricsCollector) OnRequestCancel(_ int64, statTyes []grpcapi.StatType) {
	c.mu.Lock()
	for _, statType := range statTyes {
		c.work[statType]--
		if c.work[statType] < 0 {
			c.work[statType] = 0
		}
	}
	c.mu.Unlock()
}

func (c *MetricsCollector) prepareLoadAverageResponse(averagingPeriod int64, response *grpcapi.StatsResponse) {
	if !c.statsConfig.LoadAverage {
		return
	}
	if avgStats := c.metrics.GetAverageLoadAverage(time.Duration(averagingPeriod) * time.Second); avgStats != nil {
		response.LoadAverage = converter.LoadAverageToProto(avgStats)
	}
}

func (c *MetricsCollector) prepareCPUStatsResponse(averagingPeriod int64, response *grpcapi.StatsResponse) {
	if !c.statsConfig.CPUStats {
		return
	}
	if avgStats := c.metrics.GetAverageCPUStats(time.Duration(averagingPeriod) * time.Second); avgStats != nil {
		response.CpuStats = converter.CPUStatToProto(avgStats)
	}
}

func (c *MetricsCollector) prepareDisksLoadResponse(averagingPeriod int64, response *grpcapi.StatsResponse) {
	if !c.statsConfig.DisksLoad {
		return
	}
	if avgStats := c.metrics.GetAverageDisksLoad(time.Duration(averagingPeriod) * time.Second); avgStats != nil {
		response.DisksLoad = converter.DisksLoadToProto(avgStats)
	}
}

func (c *MetricsCollector) prepareDiskStatsResponse(averagingPeriod int64, response *grpcapi.StatsResponse) {
	if !c.statsConfig.DisksStats {
		return
	}
	if stats := c.metrics.GetAverageDisksStats(time.Duration(averagingPeriod) * time.Second); stats != nil {
		response.DisksStats = converter.DiskStatsToProto(stats)
	}
}

func (c *MetricsCollector) PrepareResponse(averagingPeriod int64, statTypes []grpcapi.StatType) *grpcapi.StatsResponse {
	response := &grpcapi.StatsResponse{
		Timestamp: time.Now().Unix(),
	}

	for _, statType := range statTypes {
		switch statType {
		case grpcapi.StatType_LOAD_AVERAGE:
			c.prepareLoadAverageResponse(averagingPeriod, response)
		case grpcapi.StatType_CPU_STATS:
			c.prepareCPUStatsResponse(averagingPeriod, response)
		case grpcapi.StatType_DISKS_LOAD:
			c.prepareDisksLoadResponse(averagingPeriod, response)
		case grpcapi.StatType_DISKS_STATS:
			c.prepareDiskStatsResponse(averagingPeriod, response)
		case grpcapi.StatType_NETWORK_TOP_TALKERS:
		case grpcapi.StatType_NETWORK_CONN_STATS:
		}
	}

	return response
}
