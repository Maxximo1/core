// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bid.proto

/*
Package sonm is a generated protocol buffer package.

It is generated from these files:
	bid.proto
	capabilities.proto
	hub.proto
	insonmnia.proto
	miner.proto

It has these top-level messages:
	CPURequirements
	RAMRequirements
	DiskRequirements
	NetworkRequirements
	GPURequirements
	BidItem
	Capabilities
	CPUDevice
	RAMDevice
	GPUDevice
	ListRequest
	ListReply
	HubInfoRequest
	TaskRequirements
	HubStartTaskRequest
	HubStartTaskReply
	HubStatusMapRequest
	HubStatusRequest
	HubStatusReply
	PingRequest
	PingReply
	CPUUsage
	MemoryUsage
	NetworkUsage
	ResourceUsage
	InfoReply
	StopTaskRequest
	StopTaskReply
	TaskStatusRequest
	TaskStatusReply
	StatusMapReply
	ContainerRestartPolicy
	TaskLogsRequest
	TaskLogsChunk
	TaskResourceRequirements
	Timestamp
	MinerInfoRequest
	MinerHandshakeRequest
	MinerHandshakeReply
	MinerStartRequest
	MinerStartReply
	MinerStatusMapRequest
*/
package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// CPU requirements.
type CPURequirements struct {
	// CPU types acceptable. Empty list means any.
	Type []string `protobuf:"bytes,1,rep,name=type" json:"type,omitempty"`
	// CPU frequency minimum, measured in MHz.
	FreqMin int64 `protobuf:"varint,2,opt,name=freqMin" json:"freqMin,omitempty"`
	// CPU speed minimum, measured in some benchmark.
	SpeedMin int64 `protobuf:"varint,3,opt,name=speedMin" json:"speedMin,omitempty"`
	// CPU core count minimum.
	CoresMin int32 `protobuf:"varint,4,opt,name=coresMin" json:"coresMin,omitempty"`
}

func (m *CPURequirements) Reset()                    { *m = CPURequirements{} }
func (m *CPURequirements) String() string            { return proto.CompactTextString(m) }
func (*CPURequirements) ProtoMessage()               {}
func (*CPURequirements) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CPURequirements) GetType() []string {
	if m != nil {
		return m.Type
	}
	return nil
}

func (m *CPURequirements) GetFreqMin() int64 {
	if m != nil {
		return m.FreqMin
	}
	return 0
}

func (m *CPURequirements) GetSpeedMin() int64 {
	if m != nil {
		return m.SpeedMin
	}
	return 0
}

func (m *CPURequirements) GetCoresMin() int32 {
	if m != nil {
		return m.CoresMin
	}
	return 0
}

// RAM requirements.
type RAMRequirements struct {
	// RAM types acceptable. Empty list means any.
	Type []string `protobuf:"bytes,1,rep,name=type" json:"type,omitempty"`
	// Number of RAM bytes minimum required.
	AmountMin int64 `protobuf:"varint,2,opt,name=amountMin" json:"amountMin,omitempty"`
}

func (m *RAMRequirements) Reset()                    { *m = RAMRequirements{} }
func (m *RAMRequirements) String() string            { return proto.CompactTextString(m) }
func (*RAMRequirements) ProtoMessage()               {}
func (*RAMRequirements) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RAMRequirements) GetType() []string {
	if m != nil {
		return m.Type
	}
	return nil
}

func (m *RAMRequirements) GetAmountMin() int64 {
	if m != nil {
		return m.AmountMin
	}
	return 0
}

// Disk requirements.
type DiskRequirements struct {
	// Disk types acceptable. Empty list means any.
	Type []string `protobuf:"bytes,1,rep,name=type" json:"type,omitempty"`
	// Number of bytes minimum required.
	AmountMin int64 `protobuf:"varint,2,opt,name=amountMin" json:"amountMin,omitempty"`
	// Random read throughput minimum, measured in Kbytes/sec.
	ReadRandSpeedMin int64 `protobuf:"varint,3,opt,name=readRandSpeedMin" json:"readRandSpeedMin,omitempty"`
	// Random read throughput minimum, measured in IOPS.
	ReadRandIopsMin int64 `protobuf:"varint,4,opt,name=readRandIopsMin" json:"readRandIopsMin,omitempty"`
	// Random write throughput minimum, measured in Kbytes/sec.
	WriteRandSpeedMin int64 `protobuf:"varint,5,opt,name=writeRandSpeedMin" json:"writeRandSpeedMin,omitempty"`
	// Random write throughput minimum, measured in IOPS.
	WriteRandIopsMin int64 `protobuf:"varint,6,opt,name=writeRandIopsMin" json:"writeRandIopsMin,omitempty"`
	// I/O latency maximum, measured in microseconds.
	LatencyMax int64 `protobuf:"varint,7,opt,name=latencyMax" json:"latencyMax,omitempty"`
}

func (m *DiskRequirements) Reset()                    { *m = DiskRequirements{} }
func (m *DiskRequirements) String() string            { return proto.CompactTextString(m) }
func (*DiskRequirements) ProtoMessage()               {}
func (*DiskRequirements) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *DiskRequirements) GetType() []string {
	if m != nil {
		return m.Type
	}
	return nil
}

func (m *DiskRequirements) GetAmountMin() int64 {
	if m != nil {
		return m.AmountMin
	}
	return 0
}

func (m *DiskRequirements) GetReadRandSpeedMin() int64 {
	if m != nil {
		return m.ReadRandSpeedMin
	}
	return 0
}

func (m *DiskRequirements) GetReadRandIopsMin() int64 {
	if m != nil {
		return m.ReadRandIopsMin
	}
	return 0
}

func (m *DiskRequirements) GetWriteRandSpeedMin() int64 {
	if m != nil {
		return m.WriteRandSpeedMin
	}
	return 0
}

func (m *DiskRequirements) GetWriteRandIopsMin() int64 {
	if m != nil {
		return m.WriteRandIopsMin
	}
	return 0
}

func (m *DiskRequirements) GetLatencyMax() int64 {
	if m != nil {
		return m.LatencyMax
	}
	return 0
}

// Internet connection requirements.
type NetworkRequirements struct {
	// Outbound throughput of Internet connection minimum, measured in Kbytes/sec.
	OutSpeedMin int64 `protobuf:"varint,1,opt,name=outSpeedMin" json:"outSpeedMin,omitempty"`
	// Inbound throughput of Internet connection minimum, measured in Kbytes/sec.
	InSpeedMin int64 `protobuf:"varint,2,opt,name=inSpeedMin" json:"inSpeedMin,omitempty"`
	// Round-trip latency to some public resource maximum, measured in ms. TODO: DOUBTFUL.
	// float networkLatencyMax = 3;
	// Flag to use "white IP" for incoming connections.
	WhiteIp bool `protobuf:"varint,4,opt,name=whiteIp" json:"whiteIp,omitempty"`
}

func (m *NetworkRequirements) Reset()                    { *m = NetworkRequirements{} }
func (m *NetworkRequirements) String() string            { return proto.CompactTextString(m) }
func (*NetworkRequirements) ProtoMessage()               {}
func (*NetworkRequirements) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *NetworkRequirements) GetOutSpeedMin() int64 {
	if m != nil {
		return m.OutSpeedMin
	}
	return 0
}

func (m *NetworkRequirements) GetInSpeedMin() int64 {
	if m != nil {
		return m.InSpeedMin
	}
	return 0
}

func (m *NetworkRequirements) GetWhiteIp() bool {
	if m != nil {
		return m.WhiteIp
	}
	return false
}

type GPURequirements struct {
	// GPU types acceptable. Empty list means any.
	Type []string `protobuf:"bytes,1,rep,name=type" json:"type,omitempty"`
	// RAM minimum, measured in Gbytes. TODO: WAT?
	RamMin int64 `protobuf:"varint,2,opt,name=ramMin" json:"ramMin,omitempty"`
	// Frequency minimum, measured in MHz
	FreqMin int64 `protobuf:"varint,3,opt,name=freqMin" json:"freqMin,omitempty"`
	// Shaders count minimum.
	ShadersMin int64 `protobuf:"varint,4,opt,name=shadersMin" json:"shadersMin,omitempty"`
	// Texture mapping unit count minimum.
	TmusMin int64 `protobuf:"varint,5,opt,name=tmusMin" json:"tmusMin,omitempty"`
	// Render output unit count minimum.
	RopsMin int64 `protobuf:"varint,6,opt,name=ropsMin" json:"ropsMin,omitempty"`
}

func (m *GPURequirements) Reset()                    { *m = GPURequirements{} }
func (m *GPURequirements) String() string            { return proto.CompactTextString(m) }
func (*GPURequirements) ProtoMessage()               {}
func (*GPURequirements) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *GPURequirements) GetType() []string {
	if m != nil {
		return m.Type
	}
	return nil
}

func (m *GPURequirements) GetRamMin() int64 {
	if m != nil {
		return m.RamMin
	}
	return 0
}

func (m *GPURequirements) GetFreqMin() int64 {
	if m != nil {
		return m.FreqMin
	}
	return 0
}

func (m *GPURequirements) GetShadersMin() int64 {
	if m != nil {
		return m.ShadersMin
	}
	return 0
}

func (m *GPURequirements) GetTmusMin() int64 {
	if m != nil {
		return m.TmusMin
	}
	return 0
}

func (m *GPURequirements) GetRopsMin() int64 {
	if m != nil {
		return m.RopsMin
	}
	return 0
}

type BidItem struct {
	// When a container should be run (hour-grained).
	StartTime *Timestamp `protobuf:"bytes,1,opt,name=startTime" json:"startTime,omitempty"`
	// When a container should to be stopped (hour-grained).
	EndTime *Timestamp `protobuf:"bytes,2,opt,name=endTime" json:"endTime,omitempty"`
	// Minimal scale factor for a container.
	CountMin int32 `protobuf:"varint,3,opt,name=countMin" json:"countMin,omitempty"`
	// Maximum scale factor for a container.
	CountMax int32 `protobuf:"varint,4,opt,name=countMax" json:"countMax,omitempty"`
	// Probability of node failure maximum.
	FailureMax float32 `protobuf:"fixed32,5,opt,name=failureMax" json:"failureMax,omitempty"`
	// Node availability minimum.
	AvailabilityMin float32 `protobuf:"fixed32,6,opt,name=availabilityMin" json:"availabilityMin,omitempty"`
	// Node rating minimum.
	RatingMin float32 `protobuf:"fixed32,7,opt,name=ratingMin" json:"ratingMin,omitempty"`
	// Price per hour maximum.
	// If omitted, the minimum price for matched node is used (market price),
	// or if node also asks for market price, the price is evaluated
	// automatically (TBD).
	PriceMax float32 `protobuf:"fixed32,8,opt,name=priceMax" json:"priceMax,omitempty"`
}

func (m *BidItem) Reset()                    { *m = BidItem{} }
func (m *BidItem) String() string            { return proto.CompactTextString(m) }
func (*BidItem) ProtoMessage()               {}
func (*BidItem) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *BidItem) GetStartTime() *Timestamp {
	if m != nil {
		return m.StartTime
	}
	return nil
}

func (m *BidItem) GetEndTime() *Timestamp {
	if m != nil {
		return m.EndTime
	}
	return nil
}

func (m *BidItem) GetCountMin() int32 {
	if m != nil {
		return m.CountMin
	}
	return 0
}

func (m *BidItem) GetCountMax() int32 {
	if m != nil {
		return m.CountMax
	}
	return 0
}

func (m *BidItem) GetFailureMax() float32 {
	if m != nil {
		return m.FailureMax
	}
	return 0
}

func (m *BidItem) GetAvailabilityMin() float32 {
	if m != nil {
		return m.AvailabilityMin
	}
	return 0
}

func (m *BidItem) GetRatingMin() float32 {
	if m != nil {
		return m.RatingMin
	}
	return 0
}

func (m *BidItem) GetPriceMax() float32 {
	if m != nil {
		return m.PriceMax
	}
	return 0
}

func init() {
	proto.RegisterType((*CPURequirements)(nil), "sonm.CPURequirements")
	proto.RegisterType((*RAMRequirements)(nil), "sonm.RAMRequirements")
	proto.RegisterType((*DiskRequirements)(nil), "sonm.DiskRequirements")
	proto.RegisterType((*NetworkRequirements)(nil), "sonm.NetworkRequirements")
	proto.RegisterType((*GPURequirements)(nil), "sonm.GPURequirements")
	proto.RegisterType((*BidItem)(nil), "sonm.BidItem")
}

func init() { proto.RegisterFile("bid.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 465 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xcd, 0x8e, 0xd3, 0x30,
	0x10, 0x56, 0x92, 0x6e, 0xd3, 0xce, 0x1e, 0xb2, 0x18, 0x09, 0x55, 0x2b, 0xb4, 0xaa, 0x72, 0x2a,
	0x08, 0x7a, 0x80, 0x27, 0x80, 0x45, 0x42, 0x3d, 0x14, 0x21, 0x03, 0x0f, 0xe0, 0x36, 0xb3, 0xac,
	0x45, 0xe3, 0xa4, 0xb6, 0x43, 0xb6, 0xef, 0xc1, 0x9d, 0x07, 0xe0, 0x25, 0xd1, 0x38, 0x75, 0x7e,
	0x5a, 0x24, 0x90, 0x38, 0xb5, 0xdf, 0x8f, 0xc7, 0x33, 0xa3, 0xcf, 0x81, 0xe9, 0x46, 0x66, 0xcb,
	0x52, 0x17, 0xb6, 0x60, 0x23, 0x53, 0xa8, 0xfc, 0x3a, 0x91, 0x8a, 0x7e, 0x95, 0x14, 0x0d, 0x9d,
	0xd6, 0x90, 0xdc, 0x7e, 0xfc, 0xc2, 0x71, 0x5f, 0x49, 0x8d, 0x39, 0x2a, 0x6b, 0x18, 0x83, 0x91,
	0x3d, 0x94, 0x38, 0x0b, 0xe6, 0xd1, 0x62, 0xca, 0xdd, 0x7f, 0x36, 0x83, 0xf8, 0x4e, 0xe3, 0x7e,
	0x2d, 0xd5, 0x2c, 0x9c, 0x07, 0x8b, 0x88, 0x7b, 0xc8, 0xae, 0x61, 0x62, 0x4a, 0xc4, 0x8c, 0xa4,
	0xc8, 0x49, 0x2d, 0x26, 0x6d, 0x5b, 0x68, 0x34, 0xa4, 0x8d, 0xe6, 0xc1, 0xe2, 0x82, 0xb7, 0x38,
	0xbd, 0x85, 0x84, 0xbf, 0x59, 0xff, 0xf5, 0xe2, 0xa7, 0x30, 0x15, 0x79, 0x51, 0x29, 0xdb, 0x5d,
	0xdd, 0x11, 0xe9, 0x8f, 0x10, 0xae, 0xde, 0x49, 0xf3, 0xed, 0xff, 0xca, 0xb0, 0xe7, 0x70, 0xa5,
	0x51, 0x64, 0x5c, 0xa8, 0xec, 0xd3, 0x70, 0x96, 0x33, 0x9e, 0x2d, 0x20, 0xf1, 0xdc, 0xaa, 0x28,
	0xdb, 0xd1, 0x22, 0x7e, 0x4a, 0xb3, 0x17, 0xf0, 0xa8, 0xd6, 0xd2, 0xe2, 0xa0, 0xec, 0x85, 0xf3,
	0x9e, 0x0b, 0xd4, 0x43, 0x4b, 0xfa, 0xc2, 0xe3, 0xa6, 0x87, 0x53, 0x9e, 0xdd, 0x00, 0xec, 0x84,
	0x45, 0xb5, 0x3d, 0xac, 0xc5, 0xc3, 0x2c, 0x76, 0xae, 0x1e, 0x93, 0xee, 0xe1, 0xf1, 0x07, 0xb4,
	0x75, 0xa1, 0x87, 0x8b, 0x99, 0xc3, 0x65, 0x51, 0xd9, 0xb6, 0x95, 0xc0, 0x9d, 0xeb, 0x53, 0x54,
	0x58, 0xaa, 0xd6, 0xd0, 0xec, 0xa9, 0xc7, 0x50, 0x0c, 0xea, 0x7b, 0x69, 0x71, 0x55, 0xba, 0xa1,
	0x27, 0xdc, 0xc3, 0xf4, 0x57, 0x00, 0xc9, 0xfb, 0x7f, 0x08, 0xd2, 0x13, 0x18, 0x6b, 0x91, 0x77,
	0xd5, 0x8f, 0xa8, 0x1f, 0xb0, 0x68, 0x18, 0xb0, 0x1b, 0x00, 0x73, 0x2f, 0x32, 0xd4, 0xbd, 0x5d,
	0xf7, 0x18, 0x3a, 0x69, 0xf3, 0xca, 0x74, 0xcb, 0xf5, 0x90, 0x14, 0x3d, 0xd8, 0xa4, 0x87, 0xe9,
	0xcf, 0x10, 0xe2, 0xb7, 0x32, 0x5b, 0x59, 0xcc, 0xd9, 0x4b, 0x98, 0x1a, 0x2b, 0xb4, 0xfd, 0x2c,
	0x73, 0x74, 0x3b, 0xb9, 0x7c, 0x95, 0x2c, 0xe9, 0x91, 0x2c, 0x89, 0x31, 0x56, 0xe4, 0x25, 0xef,
	0x1c, 0xec, 0x19, 0xc4, 0xa8, 0x32, 0x67, 0x0e, 0xff, 0x6c, 0xf6, 0x7a, 0x13, 0xff, 0x63, 0xe6,
	0x22, 0x1f, 0xff, 0x63, 0xe4, 0x5a, 0x4d, 0x3c, 0x74, 0x4f, 0xa3, 0xc1, 0x34, 0xf1, 0x9d, 0x90,
	0xbb, 0x4a, 0x23, 0xa9, 0x34, 0x54, 0xc8, 0x7b, 0x0c, 0x45, 0x50, 0x7c, 0x17, 0x72, 0x27, 0x36,
	0x72, 0x27, 0xed, 0xc1, 0xcf, 0x17, 0xf2, 0x53, 0x9a, 0x62, 0xaf, 0x85, 0x95, 0xea, 0x2b, 0x79,
	0x62, 0xe7, 0xe9, 0x08, 0xea, 0xa1, 0xd4, 0x72, 0xeb, 0x6e, 0x99, 0x38, 0xb1, 0xc5, 0x9b, 0xb1,
	0xfb, 0x3c, 0xbc, 0xfe, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x73, 0x36, 0x71, 0x83, 0x42, 0x04, 0x00,
	0x00,
}
