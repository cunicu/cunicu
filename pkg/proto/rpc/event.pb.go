// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.1
// source: rpc/event.proto

package rpc

import (
	proto "cunicu.li/cunicu/pkg/proto"
	core "cunicu.li/cunicu/pkg/proto/core"
	signaling "cunicu.li/cunicu/pkg/proto/signaling"
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

type EventType int32

const (
	// Signaling Events
	EventType_BACKEND_READY     EventType = 0
	EventType_SIGNALING_MESSAGE EventType = 1
	// Core Events
	EventType_PEER_ADDED         EventType = 10
	EventType_PEER_REMOVED       EventType = 11
	EventType_PEER_MODIFIED      EventType = 12
	EventType_PEER_STATE_CHANGED EventType = 13
	EventType_INTERFACE_ADDED    EventType = 20
	EventType_INTERFACE_REMOVED  EventType = 21
	EventType_INTERFACE_MODIFIED EventType = 22
)

// Enum value maps for EventType.
var (
	EventType_name = map[int32]string{
		0:  "BACKEND_READY",
		1:  "SIGNALING_MESSAGE",
		10: "PEER_ADDED",
		11: "PEER_REMOVED",
		12: "PEER_MODIFIED",
		13: "PEER_STATE_CHANGED",
		20: "INTERFACE_ADDED",
		21: "INTERFACE_REMOVED",
		22: "INTERFACE_MODIFIED",
	}
	EventType_value = map[string]int32{
		"BACKEND_READY":      0,
		"SIGNALING_MESSAGE":  1,
		"PEER_ADDED":         10,
		"PEER_REMOVED":       11,
		"PEER_MODIFIED":      12,
		"PEER_STATE_CHANGED": 13,
		"INTERFACE_ADDED":    20,
		"INTERFACE_REMOVED":  21,
		"INTERFACE_MODIFIED": 22,
	}
)

func (x EventType) Enum() *EventType {
	p := new(EventType)
	*p = x
	return p
}

func (x EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_rpc_event_proto_enumTypes[0].Descriptor()
}

func (EventType) Type() protoreflect.EnumType {
	return &file_rpc_event_proto_enumTypes[0]
}

func (x EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EventType.Descriptor instead.
func (EventType) EnumDescriptor() ([]byte, []int) {
	return file_rpc_event_proto_rawDescGZIP(), []int{0}
}

type Event struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Type  EventType              `protobuf:"varint,1,opt,name=type,proto3,enum=cunicu.rpc.EventType" json:"type,omitempty"`
	Time  *proto.Timestamp       `protobuf:"bytes,2,opt,name=time,proto3" json:"time,omitempty"`
	// Public key of peer which triggerd the event
	Peer []byte `protobuf:"bytes,3,opt,name=peer,proto3" json:"peer,omitempty"`
	// Interface name which triggered the event
	Interface string `protobuf:"bytes,4,opt,name=interface,proto3" json:"interface,omitempty"`
	// Types that are valid to be assigned to Event:
	//
	//	*Event_BackendReady
	//	*Event_PeerStateChange
	//	*Event_PeerModified
	//	*Event_InterfaceModified
	Event         isEvent_Event `protobuf_oneof:"event"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Event) Reset() {
	*x = Event{}
	mi := &file_rpc_event_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_event_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_rpc_event_proto_rawDescGZIP(), []int{0}
}

func (x *Event) GetType() EventType {
	if x != nil {
		return x.Type
	}
	return EventType_BACKEND_READY
}

func (x *Event) GetTime() *proto.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *Event) GetPeer() []byte {
	if x != nil {
		return x.Peer
	}
	return nil
}

func (x *Event) GetInterface() string {
	if x != nil {
		return x.Interface
	}
	return ""
}

func (x *Event) GetEvent() isEvent_Event {
	if x != nil {
		return x.Event
	}
	return nil
}

func (x *Event) GetBackendReady() *SignalingBackendReadyEvent {
	if x != nil {
		if x, ok := x.Event.(*Event_BackendReady); ok {
			return x.BackendReady
		}
	}
	return nil
}

func (x *Event) GetPeerStateChange() *PeerStateChangeEvent {
	if x != nil {
		if x, ok := x.Event.(*Event_PeerStateChange); ok {
			return x.PeerStateChange
		}
	}
	return nil
}

func (x *Event) GetPeerModified() *PeerModifiedEvent {
	if x != nil {
		if x, ok := x.Event.(*Event_PeerModified); ok {
			return x.PeerModified
		}
	}
	return nil
}

func (x *Event) GetInterfaceModified() *InterfaceModifiedEvent {
	if x != nil {
		if x, ok := x.Event.(*Event_InterfaceModified); ok {
			return x.InterfaceModified
		}
	}
	return nil
}

type isEvent_Event interface {
	isEvent_Event()
}

type Event_BackendReady struct {
	BackendReady *SignalingBackendReadyEvent `protobuf:"bytes,100,opt,name=backend_ready,json=backendReady,proto3,oneof"`
}

type Event_PeerStateChange struct {
	PeerStateChange *PeerStateChangeEvent `protobuf:"bytes,121,opt,name=peer_state_change,json=peerStateChange,proto3,oneof"`
}

type Event_PeerModified struct {
	PeerModified *PeerModifiedEvent `protobuf:"bytes,122,opt,name=peer_modified,json=peerModified,proto3,oneof"`
}

type Event_InterfaceModified struct {
	InterfaceModified *InterfaceModifiedEvent `protobuf:"bytes,123,opt,name=interface_modified,json=interfaceModified,proto3,oneof"`
}

func (*Event_BackendReady) isEvent_Event() {}

func (*Event_PeerStateChange) isEvent_Event() {}

func (*Event_PeerModified) isEvent_Event() {}

func (*Event_InterfaceModified) isEvent_Event() {}

type PeerModifiedEvent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Modified      uint32                 `protobuf:"varint,1,opt,name=modified,proto3" json:"modified,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PeerModifiedEvent) Reset() {
	*x = PeerModifiedEvent{}
	mi := &file_rpc_event_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PeerModifiedEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerModifiedEvent) ProtoMessage() {}

func (x *PeerModifiedEvent) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_event_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerModifiedEvent.ProtoReflect.Descriptor instead.
func (*PeerModifiedEvent) Descriptor() ([]byte, []int) {
	return file_rpc_event_proto_rawDescGZIP(), []int{1}
}

func (x *PeerModifiedEvent) GetModified() uint32 {
	if x != nil {
		return x.Modified
	}
	return 0
}

type InterfaceModifiedEvent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Modified      uint32                 `protobuf:"varint,1,opt,name=modified,proto3" json:"modified,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InterfaceModifiedEvent) Reset() {
	*x = InterfaceModifiedEvent{}
	mi := &file_rpc_event_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InterfaceModifiedEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InterfaceModifiedEvent) ProtoMessage() {}

func (x *InterfaceModifiedEvent) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_event_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InterfaceModifiedEvent.ProtoReflect.Descriptor instead.
func (*InterfaceModifiedEvent) Descriptor() ([]byte, []int) {
	return file_rpc_event_proto_rawDescGZIP(), []int{2}
}

func (x *InterfaceModifiedEvent) GetModified() uint32 {
	if x != nil {
		return x.Modified
	}
	return 0
}

type PeerStateChangeEvent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NewState      core.PeerState         `protobuf:"varint,1,opt,name=new_state,json=newState,proto3,enum=cunicu.core.PeerState" json:"new_state,omitempty"`
	PrevState     core.PeerState         `protobuf:"varint,2,opt,name=prev_state,json=prevState,proto3,enum=cunicu.core.PeerState" json:"prev_state,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PeerStateChangeEvent) Reset() {
	*x = PeerStateChangeEvent{}
	mi := &file_rpc_event_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PeerStateChangeEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerStateChangeEvent) ProtoMessage() {}

func (x *PeerStateChangeEvent) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_event_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerStateChangeEvent.ProtoReflect.Descriptor instead.
func (*PeerStateChangeEvent) Descriptor() ([]byte, []int) {
	return file_rpc_event_proto_rawDescGZIP(), []int{3}
}

func (x *PeerStateChangeEvent) GetNewState() core.PeerState {
	if x != nil {
		return x.NewState
	}
	return core.PeerState(0)
}

func (x *PeerStateChangeEvent) GetPrevState() core.PeerState {
	if x != nil {
		return x.PrevState
	}
	return core.PeerState(0)
}

type SignalingBackendReadyEvent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          signaling.BackendType  `protobuf:"varint,1,opt,name=type,proto3,enum=cunicu.signaling.BackendType" json:"type,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SignalingBackendReadyEvent) Reset() {
	*x = SignalingBackendReadyEvent{}
	mi := &file_rpc_event_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SignalingBackendReadyEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalingBackendReadyEvent) ProtoMessage() {}

func (x *SignalingBackendReadyEvent) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_event_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalingBackendReadyEvent.ProtoReflect.Descriptor instead.
func (*SignalingBackendReadyEvent) Descriptor() ([]byte, []int) {
	return file_rpc_event_proto_rawDescGZIP(), []int{4}
}

func (x *SignalingBackendReadyEvent) GetType() signaling.BackendType {
	if x != nil {
		return x.Type
	}
	return signaling.BackendType(0)
}

var File_rpc_event_proto protoreflect.FileDescriptor

var file_rpc_event_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x72, 0x70, 0x63, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0a, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x1a, 0x0c, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x63, 0x6f, 0x72,
	0x65, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xce, 0x03, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x29, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x15, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x25, 0x0a, 0x04,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x75, 0x6e,
	0x69, 0x63, 0x75, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x74,
	0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x65, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x70, 0x65, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x66, 0x61, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x4d, 0x0a, 0x0d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64,
	0x5f, 0x72, 0x65, 0x61, 0x64, 0x79, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63,
	0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x69, 0x6e, 0x67, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x61, 0x64, 0x79, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x52,
	0x65, 0x61, 0x64, 0x79, 0x12, 0x4e, 0x0a, 0x11, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x79, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x20, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x65, 0x65,
	0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x48, 0x00, 0x52, 0x0f, 0x70, 0x65, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x43, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x12, 0x44, 0x0a, 0x0d, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x6d, 0x6f, 0x64,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x18, 0x7a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x75,
	0x6e, 0x69, 0x63, 0x75, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x4d, 0x6f, 0x64,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x70, 0x65,
	0x65, 0x72, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x12, 0x53, 0x0a, 0x12, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x18, 0x7b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e,
	0x72, 0x70, 0x63, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x4d, 0x6f, 0x64,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x11, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42,
	0x07, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x2f, 0x0a, 0x11, 0x50, 0x65, 0x65, 0x72,
	0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x22, 0x34, 0x0a, 0x16, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x22,
	0x82, 0x01, 0x0a, 0x14, 0x50, 0x65, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x43, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x33, 0x0a, 0x09, 0x6e, 0x65, 0x77, 0x5f,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x63, 0x75,
	0x6e, 0x69, 0x63, 0x75, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x52, 0x08, 0x6e, 0x65, 0x77, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x35, 0x0a,
	0x0a, 0x70, 0x72, 0x65, 0x76, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x50, 0x65, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x09, 0x70, 0x72, 0x65, 0x76, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x22, 0x4f, 0x0a, 0x1a, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e,
	0x67, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x61, 0x64, 0x79, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x12, 0x31, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x1d, 0x2e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x69, 0x6e, 0x67, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x2a, 0xc6, 0x01, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x11, 0x0a, 0x0d, 0x42, 0x41, 0x43, 0x4b, 0x45, 0x4e, 0x44, 0x5f, 0x52,
	0x45, 0x41, 0x44, 0x59, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x49, 0x47, 0x4e, 0x41, 0x4c,
	0x49, 0x4e, 0x47, 0x5f, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x10, 0x01, 0x12, 0x0e, 0x0a,
	0x0a, 0x50, 0x45, 0x45, 0x52, 0x5f, 0x41, 0x44, 0x44, 0x45, 0x44, 0x10, 0x0a, 0x12, 0x10, 0x0a,
	0x0c, 0x50, 0x45, 0x45, 0x52, 0x5f, 0x52, 0x45, 0x4d, 0x4f, 0x56, 0x45, 0x44, 0x10, 0x0b, 0x12,
	0x11, 0x0a, 0x0d, 0x50, 0x45, 0x45, 0x52, 0x5f, 0x4d, 0x4f, 0x44, 0x49, 0x46, 0x49, 0x45, 0x44,
	0x10, 0x0c, 0x12, 0x16, 0x0a, 0x12, 0x50, 0x45, 0x45, 0x52, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45,
	0x5f, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x44, 0x10, 0x0d, 0x12, 0x13, 0x0a, 0x0f, 0x49, 0x4e,
	0x54, 0x45, 0x52, 0x46, 0x41, 0x43, 0x45, 0x5f, 0x41, 0x44, 0x44, 0x45, 0x44, 0x10, 0x14, 0x12,
	0x15, 0x0a, 0x11, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x46, 0x41, 0x43, 0x45, 0x5f, 0x52, 0x45, 0x4d,
	0x4f, 0x56, 0x45, 0x44, 0x10, 0x15, 0x12, 0x16, 0x0a, 0x12, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x46,
	0x41, 0x43, 0x45, 0x5f, 0x4d, 0x4f, 0x44, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x16, 0x42, 0x20,
	0x5a, 0x1e, 0x63, 0x75, 0x6e, 0x69, 0x63, 0x75, 0x2e, 0x6c, 0x69, 0x2f, 0x63, 0x75, 0x6e, 0x69,
	0x63, 0x75, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x70, 0x63,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rpc_event_proto_rawDescOnce sync.Once
	file_rpc_event_proto_rawDescData = file_rpc_event_proto_rawDesc
)

func file_rpc_event_proto_rawDescGZIP() []byte {
	file_rpc_event_proto_rawDescOnce.Do(func() {
		file_rpc_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_rpc_event_proto_rawDescData)
	})
	return file_rpc_event_proto_rawDescData
}

var file_rpc_event_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_rpc_event_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_rpc_event_proto_goTypes = []any{
	(EventType)(0),                     // 0: cunicu.rpc.EventType
	(*Event)(nil),                      // 1: cunicu.rpc.Event
	(*PeerModifiedEvent)(nil),          // 2: cunicu.rpc.PeerModifiedEvent
	(*InterfaceModifiedEvent)(nil),     // 3: cunicu.rpc.InterfaceModifiedEvent
	(*PeerStateChangeEvent)(nil),       // 4: cunicu.rpc.PeerStateChangeEvent
	(*SignalingBackendReadyEvent)(nil), // 5: cunicu.rpc.SignalingBackendReadyEvent
	(*proto.Timestamp)(nil),            // 6: cunicu.Timestamp
	(core.PeerState)(0),                // 7: cunicu.core.PeerState
	(signaling.BackendType)(0),         // 8: cunicu.signaling.BackendType
}
var file_rpc_event_proto_depIdxs = []int32{
	0, // 0: cunicu.rpc.Event.type:type_name -> cunicu.rpc.EventType
	6, // 1: cunicu.rpc.Event.time:type_name -> cunicu.Timestamp
	5, // 2: cunicu.rpc.Event.backend_ready:type_name -> cunicu.rpc.SignalingBackendReadyEvent
	4, // 3: cunicu.rpc.Event.peer_state_change:type_name -> cunicu.rpc.PeerStateChangeEvent
	2, // 4: cunicu.rpc.Event.peer_modified:type_name -> cunicu.rpc.PeerModifiedEvent
	3, // 5: cunicu.rpc.Event.interface_modified:type_name -> cunicu.rpc.InterfaceModifiedEvent
	7, // 6: cunicu.rpc.PeerStateChangeEvent.new_state:type_name -> cunicu.core.PeerState
	7, // 7: cunicu.rpc.PeerStateChangeEvent.prev_state:type_name -> cunicu.core.PeerState
	8, // 8: cunicu.rpc.SignalingBackendReadyEvent.type:type_name -> cunicu.signaling.BackendType
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_rpc_event_proto_init() }
func file_rpc_event_proto_init() {
	if File_rpc_event_proto != nil {
		return
	}
	file_rpc_event_proto_msgTypes[0].OneofWrappers = []any{
		(*Event_BackendReady)(nil),
		(*Event_PeerStateChange)(nil),
		(*Event_PeerModified)(nil),
		(*Event_InterfaceModified)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rpc_event_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rpc_event_proto_goTypes,
		DependencyIndexes: file_rpc_event_proto_depIdxs,
		EnumInfos:         file_rpc_event_proto_enumTypes,
		MessageInfos:      file_rpc_event_proto_msgTypes,
	}.Build()
	File_rpc_event_proto = out.File
	file_rpc_event_proto_rawDesc = nil
	file_rpc_event_proto_goTypes = nil
	file_rpc_event_proto_depIdxs = nil
}
