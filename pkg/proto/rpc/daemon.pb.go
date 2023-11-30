// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: rpc/daemon.proto

package rpc

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

type ConfigValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Scalar string   `protobuf:"bytes,1,opt,name=scalar,proto3" json:"scalar,omitempty"`
	List   []string `protobuf:"bytes,2,rep,name=list,proto3" json:"list,omitempty"`
}

func (x *ConfigValue) Reset() {
	*x = ConfigValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigValue) ProtoMessage() {}

func (x *ConfigValue) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigValue.ProtoReflect.Descriptor instead.
func (*ConfigValue) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{0}
}

func (x *ConfigValue) GetScalar() string {
	if x != nil {
		return x.Scalar
	}
	return ""
}

func (x *ConfigValue) GetList() []string {
	if x != nil {
		return x.List
	}
	return nil
}

type GetStatusParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Interface string `protobuf:"bytes,1,opt,name=interface,proto3" json:"interface,omitempty"`
	Peer      []byte `protobuf:"bytes,2,opt,name=peer,proto3" json:"peer,omitempty"`
}

func (x *GetStatusParams) Reset() {
	*x = GetStatusParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStatusParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatusParams) ProtoMessage() {}

func (x *GetStatusParams) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatusParams.ProtoReflect.Descriptor instead.
func (*GetStatusParams) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{1}
}

func (x *GetStatusParams) GetInterface() string {
	if x != nil {
		return x.Interface
	}
	return ""
}

func (x *GetStatusParams) GetPeer() []byte {
	if x != nil {
		return x.Peer
	}
	return nil
}

type GetStatusResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Interfaces []*core.Interface `protobuf:"bytes,1,rep,name=interfaces,proto3" json:"interfaces,omitempty"`
}

func (x *GetStatusResp) Reset() {
	*x = GetStatusResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStatusResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatusResp) ProtoMessage() {}

func (x *GetStatusResp) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatusResp.ProtoReflect.Descriptor instead.
func (*GetStatusResp) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{2}
}

func (x *GetStatusResp) GetInterfaces() []*core.Interface {
	if x != nil {
		return x.Interfaces
	}
	return nil
}

type SetConfigParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Settings map[string]*ConfigValue `protobuf:"bytes,1,rep,name=settings,proto3" json:"settings,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *SetConfigParams) Reset() {
	*x = SetConfigParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetConfigParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetConfigParams) ProtoMessage() {}

func (x *SetConfigParams) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetConfigParams.ProtoReflect.Descriptor instead.
func (*SetConfigParams) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{3}
}

func (x *SetConfigParams) GetSettings() map[string]*ConfigValue {
	if x != nil {
		return x.Settings
	}
	return nil
}

type GetConfigParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeyFilter string `protobuf:"bytes,1,opt,name=key_filter,json=keyFilter,proto3" json:"key_filter,omitempty"`
}

func (x *GetConfigParams) Reset() {
	*x = GetConfigParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetConfigParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfigParams) ProtoMessage() {}

func (x *GetConfigParams) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetConfigParams.ProtoReflect.Descriptor instead.
func (*GetConfigParams) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{4}
}

func (x *GetConfigParams) GetKeyFilter() string {
	if x != nil {
		return x.KeyFilter
	}
	return ""
}

type GetConfigResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Settings map[string]*ConfigValue `protobuf:"bytes,1,rep,name=settings,proto3" json:"settings,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *GetConfigResp) Reset() {
	*x = GetConfigResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetConfigResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfigResp) ProtoMessage() {}

func (x *GetConfigResp) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetConfigResp.ProtoReflect.Descriptor instead.
func (*GetConfigResp) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{5}
}

func (x *GetConfigResp) GetSettings() map[string]*ConfigValue {
	if x != nil {
		return x.Settings
	}
	return nil
}

type AddPeerParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Interface string `protobuf:"bytes,1,opt,name=interface,proto3" json:"interface,omitempty"`
	PublicKey []byte `protobuf:"bytes,2,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Name      string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *AddPeerParams) Reset() {
	*x = AddPeerParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddPeerParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddPeerParams) ProtoMessage() {}

func (x *AddPeerParams) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddPeerParams.ProtoReflect.Descriptor instead.
func (*AddPeerParams) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{6}
}

func (x *AddPeerParams) GetInterface() string {
	if x != nil {
		return x.Interface
	}
	return ""
}

func (x *AddPeerParams) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *AddPeerParams) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type AddPeerResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Invitation *Invitation     `protobuf:"bytes,1,opt,name=invitation,proto3" json:"invitation,omitempty"`
	Interface  *core.Interface `protobuf:"bytes,2,opt,name=interface,proto3" json:"interface,omitempty"`
}

func (x *AddPeerResp) Reset() {
	*x = AddPeerResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddPeerResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddPeerResp) ProtoMessage() {}

func (x *AddPeerResp) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddPeerResp.ProtoReflect.Descriptor instead.
func (*AddPeerResp) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{7}
}

func (x *AddPeerResp) GetInvitation() *Invitation {
	if x != nil {
		return x.Invitation
	}
	return nil
}

func (x *AddPeerResp) GetInterface() *core.Interface {
	if x != nil {
		return x.Interface
	}
	return nil
}

type ShutdownParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Restart bool `protobuf:"varint,1,opt,name=restart,proto3" json:"restart,omitempty"`
}

func (x *ShutdownParams) Reset() {
	*x = ShutdownParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShutdownParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShutdownParams) ProtoMessage() {}

func (x *ShutdownParams) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShutdownParams.ProtoReflect.Descriptor instead.
func (*ShutdownParams) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{8}
}

func (x *ShutdownParams) GetRestart() bool {
	if x != nil {
		return x.Restart
	}
	return false
}

type GetCompletionParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cmd        []string `protobuf:"bytes,1,rep,name=cmd,proto3" json:"cmd,omitempty"`
	Args       []string `protobuf:"bytes,2,rep,name=args,proto3" json:"args,omitempty"`
	ToComplete string   `protobuf:"bytes,3,opt,name=to_complete,json=toComplete,proto3" json:"to_complete,omitempty"`
}

func (x *GetCompletionParams) Reset() {
	*x = GetCompletionParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCompletionParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCompletionParams) ProtoMessage() {}

func (x *GetCompletionParams) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCompletionParams.ProtoReflect.Descriptor instead.
func (*GetCompletionParams) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{9}
}

func (x *GetCompletionParams) GetCmd() []string {
	if x != nil {
		return x.Cmd
	}
	return nil
}

func (x *GetCompletionParams) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *GetCompletionParams) GetToComplete() string {
	if x != nil {
		return x.ToComplete
	}
	return ""
}

type GetCompletionResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Options []string `protobuf:"bytes,1,rep,name=options,proto3" json:"options,omitempty"`
	Flags   int32    `protobuf:"varint,2,opt,name=flags,proto3" json:"flags,omitempty"`
}

func (x *GetCompletionResp) Reset() {
	*x = GetCompletionResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_daemon_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCompletionResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCompletionResp) ProtoMessage() {}

func (x *GetCompletionResp) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_daemon_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCompletionResp.ProtoReflect.Descriptor instead.
func (*GetCompletionResp) Descriptor() ([]byte, []int) {
	return file_rpc_daemon_proto_rawDescGZIP(), []int{10}
}

func (x *GetCompletionResp) GetOptions() []string {
	if x != nil {
		return x.Options
	}
	return nil
}

func (x *GetCompletionResp) GetFlags() int32 {
	if x != nil {
		return x.Flags
	}
	return 0
}

var File_rpc_daemon_proto protoreflect.FileDescriptor

var file_rpc_daemon_proto_rawDesc = []byte{
	0x0a, 0x10, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0a, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x1a, 0x14,
	0x63, 0x6f, 0x72, 0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0f, 0x72, 0x70, 0x63, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x72, 0x70, 0x63, 0x2f, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x39, 0x0a, 0x0b, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x61, 0x6c,
	0x61, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x63, 0x61, 0x6c, 0x61, 0x72,
	0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04,
	0x6c, 0x69, 0x73, 0x74, 0x22, 0x43, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x66, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x65, 0x65, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x70, 0x65, 0x65, 0x72, 0x22, 0x47, 0x0a, 0x0d, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12, 0x36, 0x0a, 0x0a, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16,
	0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x52, 0x0a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63,
	0x65, 0x73, 0x22, 0xae, 0x01, 0x0a, 0x0f, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x45, 0x0a, 0x08, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e,
	0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63,
	0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x1a, 0x54, 0x0a,
	0x0d, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x2d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x22, 0x30, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x6b, 0x65, 0x79, 0x5f, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6b, 0x65, 0x79, 0x46,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x22, 0xaa, 0x01, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x12, 0x43, 0x0a, 0x08, 0x73, 0x65, 0x74, 0x74, 0x69,
	0x6e, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x63, 0x75, 0x6e, 0x69,
	0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x65, 0x73, 0x70, 0x2e, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x08, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x1a, 0x54, 0x0a, 0x0d,
	0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x2d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0x60, 0x0a, 0x0d, 0x41, 0x64, 0x64, 0x50, 0x65, 0x65, 0x72, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63,
	0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x22, 0x7b, 0x0a, 0x0b, 0x41, 0x64, 0x64, 0x50, 0x65, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x12, 0x36, 0x0a, 0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75,
	0x2e, 0x72, 0x70, 0x63, 0x2e, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x09, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16,
	0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63,
	0x65, 0x22, 0x2a, 0x0a, 0x0e, 0x53, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77, 0x6e, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x22, 0x5c, 0x0a,
	0x13, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x6d, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x6f,
	0x5f, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x74, 0x6f, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x22, 0x43, 0x0a, 0x11, 0x47,
	0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x12, 0x18, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6c,
	0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73,
	0x32, 0x8a, 0x05, 0x0a, 0x06, 0x44, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x12, 0x32, 0x0a, 0x0c, 0x47,
	0x65, 0x74, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0d, 0x2e, 0x63, 0x75,
	0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x11, 0x2e, 0x63, 0x75, 0x6e,
	0x69, 0x63, 0x75, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x00, 0x12,
	0x34, 0x0a, 0x0c, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x0d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x11,
	0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x22, 0x00, 0x30, 0x01, 0x12, 0x28, 0x0a, 0x06, 0x55, 0x6e, 0x57, 0x61, 0x69, 0x74, 0x12,
	0x0d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d,
	0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12,
	0x37, 0x0a, 0x08, 0x53, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77, 0x6e, 0x12, 0x1a, 0x2e, 0x63, 0x75,
	0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77,
	0x6e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x0d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x26, 0x0a, 0x04, 0x53, 0x79, 0x6e, 0x63,
	0x12, 0x0d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a,
	0x0d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00,
	0x12, 0x45, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1b, 0x2e,
	0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x19, 0x2e, 0x63, 0x75, 0x6e,
	0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12, 0x39, 0x0a, 0x09, 0x53, 0x65, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x1b, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70,
	0x63, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x73, 0x1a, 0x0d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x22, 0x00, 0x12, 0x45, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x1b, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x19, 0x2e, 0x63,
	0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12, 0x51, 0x0a, 0x0d, 0x47, 0x65, 0x74,
	0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x2e, 0x63, 0x75, 0x6e,
	0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x6c,
	0x65, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x1d, 0x2e, 0x63, 0x75,
	0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70,
	0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x0c,
	0x52, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x0d, 0x2e, 0x63,
	0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x63, 0x75,
	0x6e, 0x69, 0x63, 0x75, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x07,
	0x41, 0x64, 0x64, 0x50, 0x65, 0x65, 0x72, 0x12, 0x19, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75,
	0x2e, 0x72, 0x70, 0x63, 0x2e, 0x41, 0x64, 0x64, 0x50, 0x65, 0x65, 0x72, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x1a, 0x17, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x41, 0x64, 0x64, 0x50, 0x65, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x42, 0x20, 0x5a,
	0x1e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e, 0x69, 0x63,
	0x75, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x70, 0x63, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rpc_daemon_proto_rawDescOnce sync.Once
	file_rpc_daemon_proto_rawDescData = file_rpc_daemon_proto_rawDesc
)

func file_rpc_daemon_proto_rawDescGZIP() []byte {
	file_rpc_daemon_proto_rawDescOnce.Do(func() {
		file_rpc_daemon_proto_rawDescData = protoimpl.X.CompressGZIP(file_rpc_daemon_proto_rawDescData)
	})
	return file_rpc_daemon_proto_rawDescData
}

var file_rpc_daemon_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_rpc_daemon_proto_goTypes = []interface{}{
	(*ConfigValue)(nil),         // 0: cunicu.rpc.ConfigValue
	(*GetStatusParams)(nil),     // 1: cunicu.rpc.GetStatusParams
	(*GetStatusResp)(nil),       // 2: cunicu.rpc.GetStatusResp
	(*SetConfigParams)(nil),     // 3: cunicu.rpc.SetConfigParams
	(*GetConfigParams)(nil),     // 4: cunicu.rpc.GetConfigParams
	(*GetConfigResp)(nil),       // 5: cunicu.rpc.GetConfigResp
	(*AddPeerParams)(nil),       // 6: cunicu.rpc.AddPeerParams
	(*AddPeerResp)(nil),         // 7: cunicu.rpc.AddPeerResp
	(*ShutdownParams)(nil),      // 8: cunicu.rpc.ShutdownParams
	(*GetCompletionParams)(nil), // 9: cunicu.rpc.GetCompletionParams
	(*GetCompletionResp)(nil),   // 10: cunicu.rpc.GetCompletionResp
	nil,                         // 11: cunicu.rpc.SetConfigParams.SettingsEntry
	nil,                         // 12: cunicu.rpc.GetConfigResp.SettingsEntry
	(*core.Interface)(nil),      // 13: cunicu.core.Interface
	(*Invitation)(nil),          // 14: cunicu.rpc.Invitation
	(*proto.Empty)(nil),         // 15: cunicu.Empty
	(*proto.BuildInfo)(nil),     // 16: cunicu.BuildInfo
	(*Event)(nil),               // 17: cunicu.rpc.Event
}
var file_rpc_daemon_proto_depIdxs = []int32{
	13, // 0: cunicu.rpc.GetStatusResp.interfaces:type_name -> cunicu.core.Interface
	11, // 1: cunicu.rpc.SetConfigParams.settings:type_name -> cunicu.rpc.SetConfigParams.SettingsEntry
	12, // 2: cunicu.rpc.GetConfigResp.settings:type_name -> cunicu.rpc.GetConfigResp.SettingsEntry
	14, // 3: cunicu.rpc.AddPeerResp.invitation:type_name -> cunicu.rpc.Invitation
	13, // 4: cunicu.rpc.AddPeerResp.interface:type_name -> cunicu.core.Interface
	0,  // 5: cunicu.rpc.SetConfigParams.SettingsEntry.value:type_name -> cunicu.rpc.ConfigValue
	0,  // 6: cunicu.rpc.GetConfigResp.SettingsEntry.value:type_name -> cunicu.rpc.ConfigValue
	15, // 7: cunicu.rpc.Daemon.GetBuildInfo:input_type -> cunicu.Empty
	15, // 8: cunicu.rpc.Daemon.StreamEvents:input_type -> cunicu.Empty
	15, // 9: cunicu.rpc.Daemon.UnWait:input_type -> cunicu.Empty
	8,  // 10: cunicu.rpc.Daemon.Shutdown:input_type -> cunicu.rpc.ShutdownParams
	15, // 11: cunicu.rpc.Daemon.Sync:input_type -> cunicu.Empty
	1,  // 12: cunicu.rpc.Daemon.GetStatus:input_type -> cunicu.rpc.GetStatusParams
	3,  // 13: cunicu.rpc.Daemon.SetConfig:input_type -> cunicu.rpc.SetConfigParams
	4,  // 14: cunicu.rpc.Daemon.GetConfig:input_type -> cunicu.rpc.GetConfigParams
	9,  // 15: cunicu.rpc.Daemon.GetCompletion:input_type -> cunicu.rpc.GetCompletionParams
	15, // 16: cunicu.rpc.Daemon.ReloadConfig:input_type -> cunicu.Empty
	6,  // 17: cunicu.rpc.Daemon.AddPeer:input_type -> cunicu.rpc.AddPeerParams
	16, // 18: cunicu.rpc.Daemon.GetBuildInfo:output_type -> cunicu.BuildInfo
	17, // 19: cunicu.rpc.Daemon.StreamEvents:output_type -> cunicu.rpc.Event
	15, // 20: cunicu.rpc.Daemon.UnWait:output_type -> cunicu.Empty
	15, // 21: cunicu.rpc.Daemon.Shutdown:output_type -> cunicu.Empty
	15, // 22: cunicu.rpc.Daemon.Sync:output_type -> cunicu.Empty
	2,  // 23: cunicu.rpc.Daemon.GetStatus:output_type -> cunicu.rpc.GetStatusResp
	15, // 24: cunicu.rpc.Daemon.SetConfig:output_type -> cunicu.Empty
	5,  // 25: cunicu.rpc.Daemon.GetConfig:output_type -> cunicu.rpc.GetConfigResp
	10, // 26: cunicu.rpc.Daemon.GetCompletion:output_type -> cunicu.rpc.GetCompletionResp
	15, // 27: cunicu.rpc.Daemon.ReloadConfig:output_type -> cunicu.Empty
	7,  // 28: cunicu.rpc.Daemon.AddPeer:output_type -> cunicu.rpc.AddPeerResp
	18, // [18:29] is the sub-list for method output_type
	7,  // [7:18] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_rpc_daemon_proto_init() }
func file_rpc_daemon_proto_init() {
	if File_rpc_daemon_proto != nil {
		return
	}
	file_rpc_event_proto_init()
	file_rpc_invitation_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_rpc_daemon_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfigValue); i {
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
		file_rpc_daemon_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStatusParams); i {
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
		file_rpc_daemon_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStatusResp); i {
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
		file_rpc_daemon_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetConfigParams); i {
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
		file_rpc_daemon_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetConfigParams); i {
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
		file_rpc_daemon_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetConfigResp); i {
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
		file_rpc_daemon_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddPeerParams); i {
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
		file_rpc_daemon_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddPeerResp); i {
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
		file_rpc_daemon_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShutdownParams); i {
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
		file_rpc_daemon_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCompletionParams); i {
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
		file_rpc_daemon_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCompletionResp); i {
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
			RawDescriptor: file_rpc_daemon_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rpc_daemon_proto_goTypes,
		DependencyIndexes: file_rpc_daemon_proto_depIdxs,
		MessageInfos:      file_rpc_daemon_proto_msgTypes,
	}.Build()
	File_rpc_daemon_proto = out.File
	file_rpc_daemon_proto_rawDesc = nil
	file_rpc_daemon_proto_goTypes = nil
	file_rpc_daemon_proto_depIdxs = nil
}
