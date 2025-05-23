// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.1
// source: feature/pske.proto

package pske

import (
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

type PresharedKeyEstablishment struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PublicKey     []byte                 `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	CipherText    []byte                 `protobuf:"bytes,2,opt,name=cipher_text,json=cipherText,proto3" json:"cipher_text,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PresharedKeyEstablishment) Reset() {
	*x = PresharedKeyEstablishment{}
	mi := &file_feature_pske_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PresharedKeyEstablishment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PresharedKeyEstablishment) ProtoMessage() {}

func (x *PresharedKeyEstablishment) ProtoReflect() protoreflect.Message {
	mi := &file_feature_pske_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PresharedKeyEstablishment.ProtoReflect.Descriptor instead.
func (*PresharedKeyEstablishment) Descriptor() ([]byte, []int) {
	return file_feature_pske_proto_rawDescGZIP(), []int{0}
}

func (x *PresharedKeyEstablishment) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *PresharedKeyEstablishment) GetCipherText() []byte {
	if x != nil {
		return x.CipherText
	}
	return nil
}

var File_feature_pske_proto protoreflect.FileDescriptor

var file_feature_pske_proto_rawDesc = []byte{
	0x0a, 0x12, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x70, 0x73, 0x6b, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x70, 0x73, 0x6b,
	0x65, 0x22, 0x5b, 0x0a, 0x19, 0x50, 0x72, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x4b, 0x65,
	0x79, 0x45, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x1d,
	0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x1f, 0x0a,
	0x0b, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x5f, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x54, 0x65, 0x78, 0x74, 0x42, 0x29,
	0x5a, 0x27, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e, 0x69,
	0x63, 0x75, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x2f, 0x70, 0x73, 0x6b, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_feature_pske_proto_rawDescOnce sync.Once
	file_feature_pske_proto_rawDescData = file_feature_pske_proto_rawDesc
)

func file_feature_pske_proto_rawDescGZIP() []byte {
	file_feature_pske_proto_rawDescOnce.Do(func() {
		file_feature_pske_proto_rawDescData = protoimpl.X.CompressGZIP(file_feature_pske_proto_rawDescData)
	})
	return file_feature_pske_proto_rawDescData
}

var file_feature_pske_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_feature_pske_proto_goTypes = []any{
	(*PresharedKeyEstablishment)(nil), // 0: cunicu.pske.PresharedKeyEstablishment
}
var file_feature_pske_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_feature_pske_proto_init() }
func file_feature_pske_proto_init() {
	if File_feature_pske_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_feature_pske_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_feature_pske_proto_goTypes,
		DependencyIndexes: file_feature_pske_proto_depIdxs,
		MessageInfos:      file_feature_pske_proto_msgTypes,
	}.Build()
	File_feature_pske_proto = out.File
	file_feature_pske_proto_rawDesc = nil
	file_feature_pske_proto_goTypes = nil
	file_feature_pske_proto_depIdxs = nil
}
