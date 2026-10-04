package config

type DaemonConfig struct {
	RPC    RPCConfig    `mapstructure:"rpc"  json:"rpc"`
	Logger LoggerConfig `mapstructure:"logger" json:"logger"`
	Stats  StatsConfig  `mapstructure:"stats" json:"stats"`
}

func NewDaemonConfig() *DaemonConfig {
	return &DaemonConfig{}
}
