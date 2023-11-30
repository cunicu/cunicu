// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: feature/pdisc.proto

package pdisc

import (
	proto "cunicu.li/cunicu/pkg/proto"
	core "cunicu.li/cunicu/pkg/proto/core"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PeerDescriptionChange int32

const (
	PeerDescriptionChange_ADD    PeerDescriptionChange = 0
	PeerDescriptionChange_REMOVE PeerDescriptionChange = 1
	PeerDescriptionChange_UPDATE PeerDescriptionChange = 2
)

// Enum value maps for PeerDescriptionChange.
var (
	PeerDescriptionChange_name = map[int32]string{
		0: "ADD",
		1: "REMOVE",
		2: "UPDATE",
	}
	PeerDescriptionChange_value = map[string]int32{
		"ADD":    0,
		"REMOVE": 1,
		"UPDATE": 2,
	}
)

func (x PeerDescriptionChange) Enum() *PeerDescriptionChange {
	p := new(PeerDescriptionChange)
	*p = x
	return p
}

func (x PeerDescriptionChange) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PeerDescriptionChange) Descriptor() protoreflect.EnumDescriptor {
	return file_feature_pdisc_proto_enumTypes[0].Descriptor()
}

func (PeerDescriptionChange) Type() protoreflect.EnumType {
	return &file_feature_pdisc_proto_enumTypes[0]
}

func (x PeerDescriptionChange) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PeerDescriptionChange.Descriptor instead.
func (PeerDescriptionChange) EnumDescriptor() ([]byte, []int) {
	return file_feature_pdisc_proto_rawDescGZIP(), []int{0}
}

type PeerAddresses struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addresses []*core.IPAddress `protobuf:"bytes,1,rep,name=addresses,proto3" json:"addresses,omitempty"`
}

func (x *PeerAddresses) Reset() {
	*x = PeerAddresses{}
	if protoimpl.UnsafeEnabled {
		mi := &file_feature_pdisc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerAddresses) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerAddresses) ProtoMessage() {}

func (x *PeerAddresses) ProtoReflect() protoreflect.Message {
	mi := &file_feature_pdisc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerAddresses.ProtoReflect.Descriptor instead.
func (*PeerAddresses) Descriptor() ([]byte, []int) {
	return file_feature_pdisc_proto_rawDescGZIP(), []int{0}
}

func (x *PeerAddresses) GetAddresses() []*core.IPAddress {
	if x != nil {
		return x.Addresses
	}
	return nil
}

// A PeerDescription is an announcement of a peer which is distributed to
type PeerDescription struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Change PeerDescriptionChange `protobuf:"varint,1,opt,name=change,proto3,enum=cunicu.pdisc.PeerDescriptionChange" json:"change,omitempty"`
	// Hostname of the node
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Public WireGuard Curve25519 key
	PublicKey []byte `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	// A new public WireGuard Curve25519 key
	// Only valid for change == PEER_UPDATE
	PublicKeyNew []byte `protobuf:"bytes,4,opt,name=public_key_new,json=publicKeyNew,proto3" json:"public_key_new,omitempty"`
	// List of allowed IPs
	AllowedIps []string `protobuf:"bytes,5,rep,name=allowed_ips,json=allowedIps,proto3" json:"allowed_ips,omitempty"`
	// cunicu build information
	BuildInfo *proto.BuildInfo `protobuf:"bytes,6,opt,name=build_info,json=buildInfo,proto3" json:"build_info,omitempty"`
	// IP to Hostname mapping
	Hosts map[string]*PeerAddresses `protobuf:"bytes,7,rep,name=hosts,proto3" json:"hosts,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *PeerDescription) Reset() {
	*x = PeerDescription{}
	if protoimpl.UnsafeEnabled {
		mi := &file_feature_pdisc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerDescription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerDescription) ProtoMessage() {}

func (x *PeerDescription) ProtoReflect() protoreflect.Message {
	mi := &file_feature_pdisc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerDescription.ProtoReflect.Descriptor instead.
func (*PeerDescription) Descriptor() ([]byte, []int) {
	return file_feature_pdisc_proto_rawDescGZIP(), []int{1}
}

func (x *PeerDescription) GetChange() PeerDescriptionChange {
	if x != nil {
		return x.Change
	}
	return PeerDescriptionChange_ADD
}

func (x *PeerDescription) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PeerDescription) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *PeerDescription) GetPublicKeyNew() []byte {
	if x != nil {
		return x.PublicKeyNew
	}
	return nil
}

func (x *PeerDescription) GetAllowedIps() []string {
	if x != nil {
		return x.AllowedIps
	}
	return nil
}

func (x *PeerDescription) GetBuildInfo() *proto.BuildInfo {
	if x != nil {
		return x.BuildInfo
	}
	return nil
}

func (x *PeerDescription) GetHosts() map[string]*PeerAddresses {
	if x != nil {
		return x.Hosts
	}
	return nil
}

var File_feature_pdisc_proto protoreflect.FileDescriptor

var file_feature_pdisc_proto_rawDesc = []byte{
	0x0a, 0x13, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x70, 0x64,
	0x69, 0x73, 0x63, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x0e, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x6e, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x45, 0x0a, 0x0d, 0x50, 0x65, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x65, 0x73, 0x12, 0x34, 0x0a, 0x09, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x49, 0x50, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x09, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x22, 0x91, 0x03, 0x0a, 0x0f, 0x50, 0x65, 0x65,
	0x72, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3b, 0x0a, 0x06,
	0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x23, 0x2e, 0x63,
	0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x50, 0x65, 0x65, 0x72,
	0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x52, 0x06, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a,
	0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x24, 0x0a, 0x0e,
	0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x5f, 0x6e, 0x65, 0x77, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x4e,
	0x65, 0x77, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x69, 0x70,
	0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64,
	0x49, 0x70, 0x73, 0x12, 0x30, 0x0a, 0x0a, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x69, 0x6e, 0x66,
	0x6f, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x3e, 0x0a, 0x05, 0x68, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x07,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x70, 0x64,
	0x69, 0x73, 0x63, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x48, 0x6f, 0x73, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05,
	0x68, 0x6f, 0x73, 0x74, 0x73, 0x1a, 0x55, 0x0a, 0x0a, 0x48, 0x6f, 0x73, 0x74, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x31, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x70, 0x64,
	0x69, 0x73, 0x63, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65,
	0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x38, 0x0a, 0x15,
	0x50, 0x65, 0x65, 0x72, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x43,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x44, 0x44, 0x10, 0x00, 0x12, 0x0a,
	0x0a, 0x06, 0x52, 0x45, 0x4d, 0x4f, 0x56, 0x45, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x55, 0x50,
	0x44, 0x41, 0x54, 0x45, 0x10, 0x02, 0x42, 0x2a, 0x5a, 0x28, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75,
	0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x70, 0x64, 0x69,
	0x73, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_feature_pdisc_proto_rawDescOnce sync.Once
	file_feature_pdisc_proto_rawDescData = file_feature_pdisc_proto_rawDesc
)

func file_feature_pdisc_proto_rawDescGZIP() []byte {
	file_feature_pdisc_proto_rawDescOnce.Do(func() {
		file_feature_pdisc_proto_rawDescData = protoimpl.X.CompressGZIP(file_feature_pdisc_proto_rawDescData)
	})
	return file_feature_pdisc_proto_rawDescData
}

var file_feature_pdisc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_feature_pdisc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_feature_pdisc_proto_goTypes = []interface{}{
	(PeerDescriptionChange)(0), // 0: cunicu.pdisc.PeerDescriptionChange
	(*PeerAddresses)(nil),      // 1: cunicu.pdisc.PeerAddresses
	(*PeerDescription)(nil),    // 2: cunicu.pdisc.PeerDescription
	nil,                        // 3: cunicu.pdisc.PeerDescription.HostsEntry
	(*core.IPAddress)(nil),     // 4: cunicu.core.IPAddress
	(*proto.BuildInfo)(nil),    // 5: cunicu.BuildInfo
}
var file_feature_pdisc_proto_depIdxs = []int32{
	4, // 0: cunicu.pdisc.PeerAddresses.addresses:type_name -> cunicu.core.IPAddress
	0, // 1: cunicu.pdisc.PeerDescription.change:type_name -> cunicu.pdisc.PeerDescriptionChange
	5, // 2: cunicu.pdisc.PeerDescription.build_info:type_name -> cunicu.BuildInfo
	3, // 3: cunicu.pdisc.PeerDescription.hosts:type_name -> cunicu.pdisc.PeerDescription.HostsEntry
	1, // 4: cunicu.pdisc.PeerDescription.HostsEntry.value:type_name -> cunicu.pdisc.PeerAddresses
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_feature_pdisc_proto_init() }
func file_feature_pdisc_proto_init() {
	if File_feature_pdisc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_feature_pdisc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerAddresses); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_feature_pdisc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerDescription); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_feature_pdisc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_feature_pdisc_proto_goTypes,
		DependencyIndexes: file_feature_pdisc_proto_depIdxs,
		EnumInfos:         file_feature_pdisc_proto_enumTypes,
		MessageInfos:      file_feature_pdisc_proto_msgTypes,
	}.Build()
	File_feature_pdisc_proto = out.File
	file_feature_pdisc_proto_rawDesc = nil
	file_feature_pdisc_proto_goTypes = nil
	file_feature_pdisc_proto_depIdxs = nil
}
