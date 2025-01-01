// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: feature/hooks.proto

package hooks

import (
	core "cunicu.li/cunicu/pkg/proto/core"
	rpc "cunicu.li/cunicu/pkg/proto/rpc"
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

type WebHookBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      rpc.EventType   `protobuf:"varint,1,opt,name=type,proto3,enum=cunicu.rpc.EventType" json:"type,omitempty"`
	Interface *core.Interface `protobuf:"bytes,2,opt,name=interface,proto3" json:"interface,omitempty"`
	Peer      *core.Peer      `protobuf:"bytes,3,opt,name=peer,proto3" json:"peer,omitempty"`
	Modified  []string        `protobuf:"bytes,4,rep,name=modified,proto3" json:"modified,omitempty"`
}

func (x *WebHookBody) Reset() {
	*x = WebHookBody{}
	mi := &file_feature_hooks_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WebHookBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebHookBody) ProtoMessage() {}

func (x *WebHookBody) ProtoReflect() protoreflect.Message {
	mi := &file_feature_hooks_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebHookBody.ProtoReflect.Descriptor instead.
func (*WebHookBody) Descriptor() ([]byte, []int) {
	return file_feature_hooks_proto_rawDescGZIP(), []int{0}
}

func (x *WebHookBody) GetType() rpc.EventType {
	if x != nil {
		return x.Type
	}
	return rpc.EventType(0)
}

func (x *WebHookBody) GetInterface() *core.Interface {
	if x != nil {
		return x.Interface
	}
	return nil
}

func (x *WebHookBody) GetPeer() *core.Peer {
	if x != nil {
		return x.Peer
	}
	return nil
}

func (x *WebHookBody) GetModified() []string {
	if x != nil {
		return x.Modified
	}
	return nil
}

var File_feature_hooks_proto protoreflect.FileDescriptor

var file_feature_hooks_proto_rawDesc = []byte{
	0x0a, 0x13, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x68, 0x6f,
	0x6f, 0x6b, 0x73, 0x1a, 0x14, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x70, 0x65, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x72, 0x70, 0x63, 0x2f,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb1, 0x01, 0x0a, 0x0b,
	0x57, 0x65, 0x62, 0x48, 0x6f, 0x6f, 0x6b, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x29, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x63, 0x75, 0x6e, 0x69,
	0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x34, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x6e, 0x69,
	0x63, 0x75, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63,
	0x65, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x25, 0x0a, 0x04,
	0x70, 0x65, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x75, 0x6e,
	0x69, 0x63, 0x75, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x52, 0x04, 0x70,
	0x65, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42,
	0x2a, 0x5a, 0x28, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e,
	0x69, 0x63, 0x75, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x65,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_feature_hooks_proto_rawDescOnce sync.Once
	file_feature_hooks_proto_rawDescData = file_feature_hooks_proto_rawDesc
)

func file_feature_hooks_proto_rawDescGZIP() []byte {
	file_feature_hooks_proto_rawDescOnce.Do(func() {
		file_feature_hooks_proto_rawDescData = protoimpl.X.CompressGZIP(file_feature_hooks_proto_rawDescData)
	})
	return file_feature_hooks_proto_rawDescData
}

var file_feature_hooks_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_feature_hooks_proto_goTypes = []any{
	(*WebHookBody)(nil),    // 0: cunicu.hooks.WebHookBody
	(rpc.EventType)(0),     // 1: cunicu.rpc.EventType
	(*core.Interface)(nil), // 2: cunicu.core.Interface
	(*core.Peer)(nil),      // 3: cunicu.core.Peer
}
var file_feature_hooks_proto_depIdxs = []int32{
	1, // 0: cunicu.hooks.WebHookBody.type:type_name -> cunicu.rpc.EventType
	2, // 1: cunicu.hooks.WebHookBody.interface:type_name -> cunicu.core.Interface
	3, // 2: cunicu.hooks.WebHookBody.peer:type_name -> cunicu.core.Peer
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_feature_hooks_proto_init() }
func file_feature_hooks_proto_init() {
	if File_feature_hooks_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_feature_hooks_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_feature_hooks_proto_goTypes,
		DependencyIndexes: file_feature_hooks_proto_depIdxs,
		MessageInfos:      file_feature_hooks_proto_msgTypes,
	}.Build()
	File_feature_hooks_proto = out.File
	file_feature_hooks_proto_rawDesc = nil
	file_feature_hooks_proto_goTypes = nil
	file_feature_hooks_proto_depIdxs = nil
}
