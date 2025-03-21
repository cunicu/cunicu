// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.1
// source: feature/epdisc.proto

package epdisc

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

// ICE Connection state from pion/ice/ice.go
type ConnectionState int32

const (
	ConnectionState_NEW          ConnectionState = 0 // ICE agent is gathering addresses
	ConnectionState_CHECKING     ConnectionState = 1 // ICE agent has been given local and remote candidates, and is attempting to find a match
	ConnectionState_CONNECTED    ConnectionState = 2 // ICE agent has a pairing, but is still checking other pairs
	ConnectionState_COMPLETED    ConnectionState = 3 // ICE agent has finished
	ConnectionState_FAILED       ConnectionState = 4 // ICE agent never could successfully connect
	ConnectionState_DISCONNECTED ConnectionState = 5 // ICE agent connected successfully, but has entered a failed state
	ConnectionState_CLOSED       ConnectionState = 6 // ICE agent has finished and is no longer handling requests
)

// Enum value maps for ConnectionState.
var (
	ConnectionState_name = map[int32]string{
		0: "NEW",
		1: "CHECKING",
		2: "CONNECTED",
		3: "COMPLETED",
		4: "FAILED",
		5: "DISCONNECTED",
		6: "CLOSED",
	}
	ConnectionState_value = map[string]int32{
		"NEW":          0,
		"CHECKING":     1,
		"CONNECTED":    2,
		"COMPLETED":    3,
		"FAILED":       4,
		"DISCONNECTED": 5,
		"CLOSED":       6,
	}
)

func (x ConnectionState) Enum() *ConnectionState {
	p := new(ConnectionState)
	*p = x
	return p
}

func (x ConnectionState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConnectionState) Descriptor() protoreflect.EnumDescriptor {
	return file_feature_epdisc_proto_enumTypes[0].Descriptor()
}

func (ConnectionState) Type() protoreflect.EnumType {
	return &file_feature_epdisc_proto_enumTypes[0]
}

func (x ConnectionState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConnectionState.Descriptor instead.
func (ConnectionState) EnumDescriptor() ([]byte, []int) {
	return file_feature_epdisc_proto_rawDescGZIP(), []int{0}
}

type NATType int32

const (
	NATType_NONE     NATType = 0
	NATType_NFTABLES NATType = 1
)

// Enum value maps for NATType.
var (
	NATType_name = map[int32]string{
		0: "NONE",
		1: "NFTABLES",
	}
	NATType_value = map[string]int32{
		"NONE":     0,
		"NFTABLES": 1,
	}
)

func (x NATType) Enum() *NATType {
	p := new(NATType)
	*p = x
	return p
}

func (x NATType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NATType) Descriptor() protoreflect.EnumDescriptor {
	return file_feature_epdisc_proto_enumTypes[1].Descriptor()
}

func (NATType) Type() protoreflect.EnumType {
	return &file_feature_epdisc_proto_enumTypes[1]
}

func (x NATType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NATType.Descriptor instead.
func (NATType) EnumDescriptor() ([]byte, []int) {
	return file_feature_epdisc_proto_rawDescGZIP(), []int{1}
}

type ProxyType int32

const (
	ProxyType_NO_PROXY    ProxyType = 0
	ProxyType_USER_BIND   ProxyType = 1
	ProxyType_KERNEL_CONN ProxyType = 2
	ProxyType_KERNEL_NAT  ProxyType = 3
)

// Enum value maps for ProxyType.
var (
	ProxyType_name = map[int32]string{
		0: "NO_PROXY",
		1: "USER_BIND",
		2: "KERNEL_CONN",
		3: "KERNEL_NAT",
	}
	ProxyType_value = map[string]int32{
		"NO_PROXY":    0,
		"USER_BIND":   1,
		"KERNEL_CONN": 2,
		"KERNEL_NAT":  3,
	}
)

func (x ProxyType) Enum() *ProxyType {
	p := new(ProxyType)
	*p = x
	return p
}

func (x ProxyType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProxyType) Descriptor() protoreflect.EnumDescriptor {
	return file_feature_epdisc_proto_enumTypes[2].Descriptor()
}

func (ProxyType) Type() protoreflect.EnumType {
	return &file_feature_epdisc_proto_enumTypes[2]
}

func (x ProxyType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProxyType.Descriptor instead.
func (ProxyType) EnumDescriptor() ([]byte, []int) {
	return file_feature_epdisc_proto_rawDescGZIP(), []int{2}
}

type Credentials struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// ICE username fragment
	Ufrag string `protobuf:"bytes,1,opt,name=ufrag,proto3" json:"ufrag,omitempty"`
	// ICE password
	Pwd string `protobuf:"bytes,2,opt,name=pwd,proto3" json:"pwd,omitempty"`
	// Flag to indicate that the sending peer requests the credentials of the receiving peer
	NeedCreds     bool `protobuf:"varint,3,opt,name=need_creds,json=needCreds,proto3" json:"need_creds,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Credentials) Reset() {
	*x = Credentials{}
	mi := &file_feature_epdisc_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Credentials) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Credentials) ProtoMessage() {}

func (x *Credentials) ProtoReflect() protoreflect.Message {
	mi := &file_feature_epdisc_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Credentials.ProtoReflect.Descriptor instead.
func (*Credentials) Descriptor() ([]byte, []int) {
	return file_feature_epdisc_proto_rawDescGZIP(), []int{0}
}

func (x *Credentials) GetUfrag() string {
	if x != nil {
		return x.Ufrag
	}
	return ""
}

func (x *Credentials) GetPwd() string {
	if x != nil {
		return x.Pwd
	}
	return ""
}

func (x *Credentials) GetNeedCreds() bool {
	if x != nil {
		return x.NeedCreds
	}
	return false
}

type Interface struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NatType       NATType                `protobuf:"varint,1,opt,name=nat_type,json=natType,proto3,enum=cunicu.epdisc.NATType" json:"nat_type,omitempty"`
	MuxPort       uint32                 `protobuf:"varint,2,opt,name=mux_port,json=muxPort,proto3" json:"mux_port,omitempty"`
	MuxSrflxPort  uint32                 `protobuf:"varint,3,opt,name=mux_srflx_port,json=muxSrflxPort,proto3" json:"mux_srflx_port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Interface) Reset() {
	*x = Interface{}
	mi := &file_feature_epdisc_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Interface) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interface) ProtoMessage() {}

func (x *Interface) ProtoReflect() protoreflect.Message {
	mi := &file_feature_epdisc_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interface.ProtoReflect.Descriptor instead.
func (*Interface) Descriptor() ([]byte, []int) {
	return file_feature_epdisc_proto_rawDescGZIP(), []int{1}
}

func (x *Interface) GetNatType() NATType {
	if x != nil {
		return x.NatType
	}
	return NATType_NONE
}

func (x *Interface) GetMuxPort() uint32 {
	if x != nil {
		return x.MuxPort
	}
	return 0
}

func (x *Interface) GetMuxSrflxPort() uint32 {
	if x != nil {
		return x.MuxSrflxPort
	}
	return 0
}

type Peer struct {
	state                    protoimpl.MessageState `protogen:"open.v1"`
	ProxyType                ProxyType              `protobuf:"varint,1,opt,name=proxy_type,json=proxyType,proto3,enum=cunicu.epdisc.ProxyType" json:"proxy_type,omitempty"`
	SelectedCandidatePair    *CandidatePair         `protobuf:"bytes,4,opt,name=selected_candidate_pair,json=selectedCandidatePair,proto3" json:"selected_candidate_pair,omitempty"`
	LocalCandidateStats      []*CandidateStats      `protobuf:"bytes,6,rep,name=local_candidate_stats,json=localCandidateStats,proto3" json:"local_candidate_stats,omitempty"`
	RemoteCandidateStats     []*CandidateStats      `protobuf:"bytes,7,rep,name=remote_candidate_stats,json=remoteCandidateStats,proto3" json:"remote_candidate_stats,omitempty"`
	CandidatePairStats       []*CandidatePairStats  `protobuf:"bytes,8,rep,name=candidate_pair_stats,json=candidatePairStats,proto3" json:"candidate_pair_stats,omitempty"`
	LastStateChangeTimestamp *proto.Timestamp       `protobuf:"bytes,9,opt,name=last_state_change_timestamp,json=lastStateChangeTimestamp,proto3" json:"last_state_change_timestamp,omitempty"`
	Restarts                 uint32                 `protobuf:"varint,10,opt,name=restarts,proto3" json:"restarts,omitempty"`
	unknownFields            protoimpl.UnknownFields
	sizeCache                protoimpl.SizeCache
}

func (x *Peer) Reset() {
	*x = Peer{}
	mi := &file_feature_epdisc_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Peer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Peer) ProtoMessage() {}

func (x *Peer) ProtoReflect() protoreflect.Message {
	mi := &file_feature_epdisc_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Peer.ProtoReflect.Descriptor instead.
func (*Peer) Descriptor() ([]byte, []int) {
	return file_feature_epdisc_proto_rawDescGZIP(), []int{2}
}

func (x *Peer) GetProxyType() ProxyType {
	if x != nil {
		return x.ProxyType
	}
	return ProxyType_NO_PROXY
}

func (x *Peer) GetSelectedCandidatePair() *CandidatePair {
	if x != nil {
		return x.SelectedCandidatePair
	}
	return nil
}

func (x *Peer) GetLocalCandidateStats() []*CandidateStats {
	if x != nil {
		return x.LocalCandidateStats
	}
	return nil
}

func (x *Peer) GetRemoteCandidateStats() []*CandidateStats {
	if x != nil {
		return x.RemoteCandidateStats
	}
	return nil
}

func (x *Peer) GetCandidatePairStats() []*CandidatePairStats {
	if x != nil {
		return x.CandidatePairStats
	}
	return nil
}

func (x *Peer) GetLastStateChangeTimestamp() *proto.Timestamp {
	if x != nil {
		return x.LastStateChangeTimestamp
	}
	return nil
}

func (x *Peer) GetRestarts() uint32 {
	if x != nil {
		return x.Restarts
	}
	return 0
}

var File_feature_epdisc_proto protoreflect.FileDescriptor

var file_feature_epdisc_proto_rawDesc = []byte{
	0x0a, 0x14, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x65,
	0x70, 0x64, 0x69, 0x73, 0x63, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2f, 0x65, 0x70, 0x64,
	0x69, 0x73, 0x63, 0x5f, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x54, 0x0a, 0x0b, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61,
	0x6c, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x75, 0x66, 0x72, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x75, 0x66, 0x72, 0x61, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x77, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x70, 0x77, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x6e, 0x65,
	0x65, 0x64, 0x5f, 0x63, 0x72, 0x65, 0x64, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09,
	0x6e, 0x65, 0x65, 0x64, 0x43, 0x72, 0x65, 0x64, 0x73, 0x22, 0x7f, 0x0a, 0x09, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x31, 0x0a, 0x08, 0x6e, 0x61, 0x74, 0x5f, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63,
	0x75, 0x2e, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x4e, 0x41, 0x54, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x07, 0x6e, 0x61, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x6d, 0x75, 0x78,
	0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x6d, 0x75, 0x78,
	0x50, 0x6f, 0x72, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x6d, 0x75, 0x78, 0x5f, 0x73, 0x72, 0x66, 0x6c,
	0x78, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x6d, 0x75,
	0x78, 0x53, 0x72, 0x66, 0x6c, 0x78, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x80, 0x04, 0x0a, 0x04, 0x50,
	0x65, 0x65, 0x72, 0x12, 0x37, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75,
	0x2e, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x54, 0x0a, 0x17,
	0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x65, 0x64, 0x5f, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x5f, 0x70, 0x61, 0x69, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x43, 0x61,
	0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x61, 0x69, 0x72, 0x52, 0x15, 0x73, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x65, 0x64, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x61,
	0x69, 0x72, 0x12, 0x51, 0x0a, 0x15, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x63, 0x61, 0x6e, 0x64,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x65, 0x70, 0x64, 0x69, 0x73,
	0x63, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73,
	0x52, 0x13, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x53, 0x0a, 0x16, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x5f,
	0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x18,
	0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x65,
	0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x52, 0x14, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x43, 0x61, 0x6e, 0x64,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x53, 0x0a, 0x14, 0x63, 0x61,
	0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x73, 0x74, 0x61,
	0x74, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63,
	0x75, 0x2e, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x50, 0x61, 0x69, 0x72, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x12, 0x63, 0x61, 0x6e,
	0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x61, 0x69, 0x72, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12,
	0x50, 0x0a, 0x1b, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x18, 0x6c, 0x61, 0x73, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x73, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x08, 0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x73, 0x2a, 0x70, 0x0a,
	0x0f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x12, 0x07, 0x0a, 0x03, 0x4e, 0x45, 0x57, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x48, 0x45,
	0x43, 0x4b, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4e, 0x4e, 0x45,
	0x43, 0x54, 0x45, 0x44, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45,
	0x54, 0x45, 0x44, 0x10, 0x03, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10,
	0x04, 0x12, 0x10, 0x0a, 0x0c, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x45,
	0x44, 0x10, 0x05, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x4c, 0x4f, 0x53, 0x45, 0x44, 0x10, 0x06, 0x2a,
	0x21, 0x0a, 0x07, 0x4e, 0x41, 0x54, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x4f,
	0x4e, 0x45, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x4e, 0x46, 0x54, 0x41, 0x42, 0x4c, 0x45, 0x53,
	0x10, 0x01, 0x2a, 0x49, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x0c, 0x0a, 0x08, 0x4e, 0x4f, 0x5f, 0x50, 0x52, 0x4f, 0x58, 0x59, 0x10, 0x00, 0x12, 0x0d, 0x0a,
	0x09, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x42, 0x49, 0x4e, 0x44, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b,
	0x4b, 0x45, 0x52, 0x4e, 0x45, 0x4c, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x10, 0x02, 0x12, 0x0e, 0x0a,
	0x0a, 0x4b, 0x45, 0x52, 0x4e, 0x45, 0x4c, 0x5f, 0x4e, 0x41, 0x54, 0x10, 0x03, 0x42, 0x2b, 0x5a,
	0x29, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e, 0x69, 0x63,
	0x75, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x65, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x2f, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_feature_epdisc_proto_rawDescOnce sync.Once
	file_feature_epdisc_proto_rawDescData = file_feature_epdisc_proto_rawDesc
)

func file_feature_epdisc_proto_rawDescGZIP() []byte {
	file_feature_epdisc_proto_rawDescOnce.Do(func() {
		file_feature_epdisc_proto_rawDescData = protoimpl.X.CompressGZIP(file_feature_epdisc_proto_rawDescData)
	})
	return file_feature_epdisc_proto_rawDescData
}

var file_feature_epdisc_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_feature_epdisc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_feature_epdisc_proto_goTypes = []any{
	(ConnectionState)(0),       // 0: cunicu.epdisc.ConnectionState
	(NATType)(0),               // 1: cunicu.epdisc.NATType
	(ProxyType)(0),             // 2: cunicu.epdisc.ProxyType
	(*Credentials)(nil),        // 3: cunicu.epdisc.Credentials
	(*Interface)(nil),          // 4: cunicu.epdisc.Interface
	(*Peer)(nil),               // 5: cunicu.epdisc.Peer
	(*CandidatePair)(nil),      // 6: cunicu.epdisc.CandidatePair
	(*CandidateStats)(nil),     // 7: cunicu.epdisc.CandidateStats
	(*CandidatePairStats)(nil), // 8: cunicu.epdisc.CandidatePairStats
	(*proto.Timestamp)(nil),    // 9: cunicu.Timestamp
}
var file_feature_epdisc_proto_depIdxs = []int32{
	1, // 0: cunicu.epdisc.Interface.nat_type:type_name -> cunicu.epdisc.NATType
	2, // 1: cunicu.epdisc.Peer.proxy_type:type_name -> cunicu.epdisc.ProxyType
	6, // 2: cunicu.epdisc.Peer.selected_candidate_pair:type_name -> cunicu.epdisc.CandidatePair
	7, // 3: cunicu.epdisc.Peer.local_candidate_stats:type_name -> cunicu.epdisc.CandidateStats
	7, // 4: cunicu.epdisc.Peer.remote_candidate_stats:type_name -> cunicu.epdisc.CandidateStats
	8, // 5: cunicu.epdisc.Peer.candidate_pair_stats:type_name -> cunicu.epdisc.CandidatePairStats
	9, // 6: cunicu.epdisc.Peer.last_state_change_timestamp:type_name -> cunicu.Timestamp
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_feature_epdisc_proto_init() }
func file_feature_epdisc_proto_init() {
	if File_feature_epdisc_proto != nil {
		return
	}
	file_feature_epdisc_candidate_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_feature_epdisc_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_feature_epdisc_proto_goTypes,
		DependencyIndexes: file_feature_epdisc_proto_depIdxs,
		EnumInfos:         file_feature_epdisc_proto_enumTypes,
		MessageInfos:      file_feature_epdisc_proto_msgTypes,
	}.Build()
	File_feature_epdisc_proto = out.File
	file_feature_epdisc_proto_rawDesc = nil
	file_feature_epdisc_proto_goTypes = nil
	file_feature_epdisc_proto_depIdxs = nil
}
