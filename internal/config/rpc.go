package config

type RPCConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port uint16 `mapstructure:"port" json:"port"`
}
