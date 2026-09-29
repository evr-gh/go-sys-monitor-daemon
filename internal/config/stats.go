package config

type StatsConfig struct {
	AveragingPeriodLimit int64 `mapstructure:"averaging_period_limit" json:"averagingPeriodLimit"`
	LoadAverage          bool  `mapstructure:"load_average" json:"loadAverage"`
	CPUStats             bool  `mapstructure:"cpu_stats" json:"cpuStats"`
	DisksLoad            bool  `mapstructure:"disks_load" json:"disksLoad"`
	DisksStats           bool  `mapstructure:"disks_starts" json:"disksStats"`
	NetworkTopTalkers    bool  `mapstructure:"network_top_talkers" json:"networkTopTalkers"`
	NetworkConnStats     bool  `mapstructure:"network_conn_stats" json:"networkConnStats"`
}
