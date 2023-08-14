// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.19.6
// source: signaling/relay.proto

package signaling

import (
	proto "cunicu.li/cunicu/pkg/proto"
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

type RelayInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url      string           `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Username string           `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Password string           `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	Expires  *proto.Timestamp `protobuf:"bytes,4,opt,name=expires,proto3" json:"expires,omitempty"`
}

func (x *RelayInfo) Reset() {
	*x = RelayInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_relay_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RelayInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RelayInfo) ProtoMessage() {}

func (x *RelayInfo) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_relay_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RelayInfo.ProtoReflect.Descriptor instead.
func (*RelayInfo) Descriptor() ([]byte, []int) {
	return file_signaling_relay_proto_rawDescGZIP(), []int{0}
}

func (x *RelayInfo) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *RelayInfo) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *RelayInfo) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *RelayInfo) GetExpires() *proto.Timestamp {
	if x != nil {
		return x.Expires
	}
	return nil
}

type GetRelaysParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Public key of peer which requestes the credentials
	PublicKey []byte `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (x *GetRelaysParams) Reset() {
	*x = GetRelaysParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_relay_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRelaysParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRelaysParams) ProtoMessage() {}

func (x *GetRelaysParams) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_relay_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRelaysParams.ProtoReflect.Descriptor instead.
func (*GetRelaysParams) Descriptor() ([]byte, []int) {
	return file_signaling_relay_proto_rawDescGZIP(), []int{1}
}

func (x *GetRelaysParams) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

type GetRelaysResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Relays []*RelayInfo `protobuf:"bytes,1,rep,name=relays,proto3" json:"relays,omitempty"`
}

func (x *GetRelaysResp) Reset() {
	*x = GetRelaysResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_relay_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRelaysResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRelaysResp) ProtoMessage() {}

func (x *GetRelaysResp) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_relay_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRelaysResp.ProtoReflect.Descriptor instead.
func (*GetRelaysResp) Descriptor() ([]byte, []int) {
	return file_signaling_relay_proto_rawDescGZIP(), []int{2}
}

func (x *GetRelaysResp) GetRelays() []*RelayInfo {
	if x != nil {
		return x.Relays
	}
	return nil
}

var File_signaling_relay_proto protoreflect.FileDescriptor

var file_signaling_relay_proto_rawDesc = []byte{
	0x0a, 0x15, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2f, 0x72, 0x65, 0x6c, 0x61,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x82, 0x01, 0x0a, 0x09, 0x52, 0x65, 0x6c, 0x61,
	0x79, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12,
	0x2b, 0x0a, 0x07, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x07, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x22, 0x30, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x73, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12,
	0x1d, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x22, 0x44,
	0x0a, 0x0d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12,
	0x33, 0x0a, 0x06, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1b, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69,
	0x6e, 0x67, 0x2e, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x06, 0x72, 0x65,
	0x6c, 0x61, 0x79, 0x73, 0x32, 0x62, 0x0a, 0x0d, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x12, 0x51, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x52, 0x65, 0x6c, 0x61,
	0x79, 0x73, 0x12, 0x21, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x73, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x1f, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x6c, 0x61,
	0x79, 0x73, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x42, 0x26, 0x5a, 0x24, 0x63, 0x75, 0x6e, 0x69,
	0x63, 0x75, 0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_signaling_relay_proto_rawDescOnce sync.Once
	file_signaling_relay_proto_rawDescData = file_signaling_relay_proto_rawDesc
)

func file_signaling_relay_proto_rawDescGZIP() []byte {
	file_signaling_relay_proto_rawDescOnce.Do(func() {
		file_signaling_relay_proto_rawDescData = protoimpl.X.CompressGZIP(file_signaling_relay_proto_rawDescData)
	})
	return file_signaling_relay_proto_rawDescData
}

var file_signaling_relay_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_signaling_relay_proto_goTypes = []interface{}{
	(*RelayInfo)(nil),       // 0: cunicu.signaling.RelayInfo
	(*GetRelaysParams)(nil), // 1: cunicu.signaling.GetRelaysParams
	(*GetRelaysResp)(nil),   // 2: cunicu.signaling.GetRelaysResp
	(*proto.Timestamp)(nil), // 3: cunicu.Timestamp
}
var file_signaling_relay_proto_depIdxs = []int32{
	3, // 0: cunicu.signaling.RelayInfo.expires:type_name -> cunicu.Timestamp
	0, // 1: cunicu.signaling.GetRelaysResp.relays:type_name -> cunicu.signaling.RelayInfo
	1, // 2: cunicu.signaling.RelayRegistry.GetRelays:input_type -> cunicu.signaling.GetRelaysParams
	2, // 3: cunicu.signaling.RelayRegistry.GetRelays:output_type -> cunicu.signaling.GetRelaysResp
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_signaling_relay_proto_init() }
func file_signaling_relay_proto_init() {
	if File_signaling_relay_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_signaling_relay_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RelayInfo); i {
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
		file_signaling_relay_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRelaysParams); i {
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
		file_signaling_relay_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRelaysResp); i {
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
			RawDescriptor: file_signaling_relay_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_signaling_relay_proto_goTypes,
		DependencyIndexes: file_signaling_relay_proto_depIdxs,
		MessageInfos:      file_signaling_relay_proto_msgTypes,
	}.Build()
	File_signaling_relay_proto = out.File
	file_signaling_relay_proto_rawDesc = nil
	file_signaling_relay_proto_goTypes = nil
	file_signaling_relay_proto_depIdxs = nil
}
