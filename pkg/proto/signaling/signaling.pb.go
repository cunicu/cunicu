// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.6.1
// source: signaling/signaling.proto

package signaling

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	proto "riasc.eu/wice/pkg/proto"
	epdisc "riasc.eu/wice/pkg/proto/feat/epdisc"
	pdisc "riasc.eu/wice/pkg/proto/feat/pdisc"
	pske "riasc.eu/wice/pkg/proto/feat/pske"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BackendType int32

const (
	BackendType_MULTI     BackendType = 0
	BackendType_P2P       BackendType = 1
	BackendType_K8S       BackendType = 2
	BackendType_GRPC      BackendType = 3
	BackendType_INPROCESS BackendType = 4
)

// Enum value maps for BackendType.
var (
	BackendType_name = map[int32]string{
		0: "MULTI",
		1: "P2P",
		2: "K8S",
		3: "GRPC",
		4: "INPROCESS",
	}
	BackendType_value = map[string]int32{
		"MULTI":     0,
		"P2P":       1,
		"K8S":       2,
		"GRPC":      3,
		"INPROCESS": 4,
	}
)

func (x BackendType) Enum() *BackendType {
	p := new(BackendType)
	*p = x
	return p
}

func (x BackendType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BackendType) Descriptor() protoreflect.EnumDescriptor {
	return file_signaling_signaling_proto_enumTypes[0].Descriptor()
}

func (BackendType) Type() protoreflect.EnumType {
	return &file_signaling_signaling_proto_enumTypes[0]
}

func (x BackendType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BackendType.Descriptor instead.
func (BackendType) EnumDescriptor() ([]byte, []int) {
	return file_signaling_signaling_proto_rawDescGZIP(), []int{0}
}

type Envelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sender    []byte            `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Recipient []byte            `protobuf:"bytes,2,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Contents  *EncryptedMessage `protobuf:"bytes,3,opt,name=contents,proto3" json:"contents,omitempty"` // of type SignalingMessage
}

func (x *Envelope) Reset() {
	*x = Envelope{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_signaling_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Envelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Envelope) ProtoMessage() {}

func (x *Envelope) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_signaling_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Envelope.ProtoReflect.Descriptor instead.
func (*Envelope) Descriptor() ([]byte, []int) {
	return file_signaling_signaling_proto_rawDescGZIP(), []int{0}
}

func (x *Envelope) GetSender() []byte {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *Envelope) GetRecipient() []byte {
	if x != nil {
		return x.Recipient
	}
	return nil
}

func (x *Envelope) GetContents() *EncryptedMessage {
	if x != nil {
		return x.Contents
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Credentials *epdisc.Credentials             `protobuf:"bytes,1,opt,name=credentials,proto3" json:"credentials,omitempty"`
	Candidate   *epdisc.Candidate               `protobuf:"bytes,2,opt,name=candidate,proto3" json:"candidate,omitempty"`
	Peer        *pdisc.PeerDescription          `protobuf:"bytes,3,opt,name=peer,proto3" json:"peer,omitempty"`
	Pske        *pske.PresharedKeyEstablishment `protobuf:"bytes,4,opt,name=pske,proto3" json:"pske,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_signaling_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_signaling_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_signaling_signaling_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetCredentials() *epdisc.Credentials {
	if x != nil {
		return x.Credentials
	}
	return nil
}

func (x *Message) GetCandidate() *epdisc.Candidate {
	if x != nil {
		return x.Candidate
	}
	return nil
}

func (x *Message) GetPeer() *pdisc.PeerDescription {
	if x != nil {
		return x.Peer
	}
	return nil
}

func (x *Message) GetPske() *pske.PresharedKeyEstablishment {
	if x != nil {
		return x.Pske
	}
	return nil
}

// A container for an encrypted protobuf message
type EncryptedMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Body  []byte `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	Nonce []byte `protobuf:"bytes,2,opt,name=nonce,proto3" json:"nonce,omitempty"`
}

func (x *EncryptedMessage) Reset() {
	*x = EncryptedMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_signaling_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EncryptedMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncryptedMessage) ProtoMessage() {}

func (x *EncryptedMessage) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_signaling_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EncryptedMessage.ProtoReflect.Descriptor instead.
func (*EncryptedMessage) Descriptor() ([]byte, []int) {
	return file_signaling_signaling_proto_rawDescGZIP(), []int{2}
}

func (x *EncryptedMessage) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *EncryptedMessage) GetNonce() []byte {
	if x != nil {
		return x.Nonce
	}
	return nil
}

type SubscribeParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *SubscribeParams) Reset() {
	*x = SubscribeParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signaling_signaling_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeParams) ProtoMessage() {}

func (x *SubscribeParams) ProtoReflect() protoreflect.Message {
	mi := &file_signaling_signaling_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeParams.ProtoReflect.Descriptor instead.
func (*SubscribeParams) Descriptor() ([]byte, []int) {
	return file_signaling_signaling_proto_rawDescGZIP(), []int{3}
}

func (x *SubscribeParams) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

var File_signaling_signaling_proto protoreflect.FileDescriptor

var file_signaling_signaling_proto_rawDesc = []byte{
	0x0a, 0x19, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2f, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x77, 0x69, 0x63,
	0x65, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x1a, 0x0c, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x66, 0x65, 0x61, 0x74, 0x2f,
	0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x66, 0x65, 0x61,
	0x74, 0x2f, 0x70, 0x73, 0x6b, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x66, 0x65,
	0x61, 0x74, 0x2f, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1b, 0x66, 0x65, 0x61, 0x74, 0x2f, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x5f, 0x63, 0x61, 0x6e,
	0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7e, 0x0a, 0x08,
	0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x72, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x3c,
	0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x77, 0x69, 0x63, 0x65, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e,
	0x67, 0x2e, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xe6, 0x01, 0x0a,
	0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x3a, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x77, 0x69, 0x63, 0x65, 0x2e, 0x65, 0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x43, 0x72, 0x65, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x61, 0x6c, 0x73, 0x12, 0x34, 0x0a, 0x09, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x77, 0x69, 0x63, 0x65, 0x2e, 0x65,
	0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x09, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x12, 0x2f, 0x0a, 0x04, 0x70, 0x65,
	0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x77, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x64, 0x69, 0x73, 0x63, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x04, 0x70, 0x65, 0x65, 0x72, 0x12, 0x38, 0x0a, 0x04, 0x70,
	0x73, 0x6b, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x77, 0x69, 0x63, 0x65,
	0x2e, 0x70, 0x73, 0x6b, 0x65, 0x2e, 0x50, 0x72, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x4b,
	0x65, 0x79, 0x45, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52,
	0x04, 0x70, 0x73, 0x6b, 0x65, 0x22, 0x3c, 0x0a, 0x10, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x14, 0x0a,
	0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x6e, 0x6f,
	0x6e, 0x63, 0x65, 0x22, 0x23, 0x0a, 0x0f, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65,
	0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x2a, 0x43, 0x0a, 0x0b, 0x42, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x4d, 0x55, 0x4c, 0x54, 0x49,
	0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x50, 0x32, 0x50, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x4b,
	0x38, 0x53, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x47, 0x52, 0x50, 0x43, 0x10, 0x03, 0x12, 0x0d,
	0x0a, 0x09, 0x49, 0x4e, 0x50, 0x52, 0x4f, 0x43, 0x45, 0x53, 0x53, 0x10, 0x04, 0x32, 0xbb, 0x01,
	0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x12, 0x2e, 0x0a, 0x0c, 0x47,
	0x65, 0x74, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0b, 0x2e, 0x77, 0x69,
	0x63, 0x65, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0f, 0x2e, 0x77, 0x69, 0x63, 0x65, 0x2e,
	0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x00, 0x12, 0x4a, 0x0a, 0x09, 0x53,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x12, 0x1f, 0x2e, 0x77, 0x69, 0x63, 0x65, 0x2e,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72,
	0x69, 0x62, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x18, 0x2e, 0x77, 0x69, 0x63, 0x65,
	0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6e, 0x76, 0x65, 0x6c,
	0x6f, 0x70, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x32, 0x0a, 0x07, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x73, 0x68, 0x12, 0x18, 0x2e, 0x77, 0x69, 0x63, 0x65, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x1a, 0x0b, 0x2e, 0x77,
	0x69, 0x63, 0x65, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x23, 0x5a, 0x21, 0x72,
	0x69, 0x61, 0x73, 0x63, 0x2e, 0x65, 0x75, 0x2f, 0x77, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_signaling_signaling_proto_rawDescOnce sync.Once
	file_signaling_signaling_proto_rawDescData = file_signaling_signaling_proto_rawDesc
)

func file_signaling_signaling_proto_rawDescGZIP() []byte {
	file_signaling_signaling_proto_rawDescOnce.Do(func() {
		file_signaling_signaling_proto_rawDescData = protoimpl.X.CompressGZIP(file_signaling_signaling_proto_rawDescData)
	})
	return file_signaling_signaling_proto_rawDescData
}

var file_signaling_signaling_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_signaling_signaling_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_signaling_signaling_proto_goTypes = []interface{}{
	(BackendType)(0),                       // 0: wice.signaling.BackendType
	(*Envelope)(nil),                       // 1: wice.signaling.Envelope
	(*Message)(nil),                        // 2: wice.signaling.Message
	(*EncryptedMessage)(nil),               // 3: wice.signaling.EncryptedMessage
	(*SubscribeParams)(nil),                // 4: wice.signaling.SubscribeParams
	(*epdisc.Credentials)(nil),             // 5: wice.epdisc.Credentials
	(*epdisc.Candidate)(nil),               // 6: wice.epdisc.Candidate
	(*pdisc.PeerDescription)(nil),          // 7: wice.pdisc.PeerDescription
	(*pske.PresharedKeyEstablishment)(nil), // 8: wice.pske.PresharedKeyEstablishment
	(*proto.Empty)(nil),                    // 9: wice.Empty
	(*proto.BuildInfo)(nil),                // 10: wice.BuildInfo
}
var file_signaling_signaling_proto_depIdxs = []int32{
	3,  // 0: wice.signaling.Envelope.contents:type_name -> wice.signaling.EncryptedMessage
	5,  // 1: wice.signaling.Message.credentials:type_name -> wice.epdisc.Credentials
	6,  // 2: wice.signaling.Message.candidate:type_name -> wice.epdisc.Candidate
	7,  // 3: wice.signaling.Message.peer:type_name -> wice.pdisc.PeerDescription
	8,  // 4: wice.signaling.Message.pske:type_name -> wice.pske.PresharedKeyEstablishment
	9,  // 5: wice.signaling.Signaling.GetBuildInfo:input_type -> wice.Empty
	4,  // 6: wice.signaling.Signaling.Subscribe:input_type -> wice.signaling.SubscribeParams
	1,  // 7: wice.signaling.Signaling.Publish:input_type -> wice.signaling.Envelope
	10, // 8: wice.signaling.Signaling.GetBuildInfo:output_type -> wice.BuildInfo
	1,  // 9: wice.signaling.Signaling.Subscribe:output_type -> wice.signaling.Envelope
	9,  // 10: wice.signaling.Signaling.Publish:output_type -> wice.Empty
	8,  // [8:11] is the sub-list for method output_type
	5,  // [5:8] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_signaling_signaling_proto_init() }
func file_signaling_signaling_proto_init() {
	if File_signaling_signaling_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_signaling_signaling_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Envelope); i {
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
		file_signaling_signaling_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
		file_signaling_signaling_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EncryptedMessage); i {
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
		file_signaling_signaling_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeParams); i {
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
			RawDescriptor: file_signaling_signaling_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_signaling_signaling_proto_goTypes,
		DependencyIndexes: file_signaling_signaling_proto_depIdxs,
		EnumInfos:         file_signaling_signaling_proto_enumTypes,
		MessageInfos:      file_signaling_signaling_proto_msgTypes,
	}.Build()
	File_signaling_signaling_proto = out.File
	file_signaling_signaling_proto_rawDesc = nil
	file_signaling_signaling_proto_goTypes = nil
	file_signaling_signaling_proto_depIdxs = nil
}
