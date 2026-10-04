package models

type CPUStat struct {
	UserMode   float64 `protobuf:"fixed64,1,opt,name=user_mode,proto3" json:"userMode"`
	SystemMode float64 `protobuf:"fixed64,2,opt,name=system_mode,proto3" json:"systemMode"`
	Idle       float64 `protobuf:"fixed64,3,opt,name=idle,proto3" json:"idle"`
}
