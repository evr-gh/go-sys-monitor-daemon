package models

type LoadAverage struct {
	Load1Min  float64 `protobuf:"fixed64,1,opt,name=load1min,proto3" json:"load1min"`
	Load5Min  float64 `protobuf:"fixed64,2,opt,name=load5min,proto3" json:"load5min"`
	Load15Min float64 `protobuf:"fixed64,3,opt,name=load15min,proto3" json:"load15min"`
}
