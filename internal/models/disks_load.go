package models

type DisksLoad struct {
	DisksLoad []DiskLoad `protobuf:"bytes,1,rep,name=disks_load,proto3" json:"disksLoad"`
}

type DiskLoad struct {
	Name     string  `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	Tps      float64 `protobuf:"fixed64,1,opt,name=tps,proto3" json:"tps"`
	KpsRead  float64 `protobuf:"fixed64,2,opt,name=kps_read,proto3" json:"kpsRead"`
	KpsWrite float64 `protobuf:"fixed64,2,opt,name=kps_write,proto3" json:"kpsWrite"`
}
