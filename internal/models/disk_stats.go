package models

type DiskStats struct {
	DiskStats []DiskStat `protobuf:"bytes,1,rep,name=disk_stats,proto3" json:"diskStats"`
}

type DiskStat struct {
	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	MBUsagePct uint32 `protobuf:"bytes,2,opt,name=mb_usage_pct,proto3" json:"mbUsagePct"`
	InUsagePct uint32 `protobuf:"bytes,3,opt,name=in_usage_pct,proto3" json:"inUsagePct"`
}
