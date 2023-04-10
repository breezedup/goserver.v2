// Code generated by protoc-gen-go.
// source: protocol/broadcast.proto
// DO NOT EDIT!

package protocol

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type SSPacketBroadcast struct {
	SessParam        *BCSessionUnion `protobuf:"bytes,1,req" json:"SessParam,omitempty"`
	PacketId         *int32          `protobuf:"varint,2,req" json:"PacketId,omitempty"`
	Data             []byte          `protobuf:"bytes,3,req" json:"Data,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *SSPacketBroadcast) Reset()         { *m = SSPacketBroadcast{} }
func (m *SSPacketBroadcast) String() string { return proto.CompactTextString(m) }
func (*SSPacketBroadcast) ProtoMessage()    {}

func (m *SSPacketBroadcast) GetSessParam() *BCSessionUnion {
	if m != nil {
		return m.SessParam
	}
	return nil
}

func (m *SSPacketBroadcast) GetPacketId() int32 {
	if m != nil && m.PacketId != nil {
		return *m.PacketId
	}
	return 0
}

func (m *SSPacketBroadcast) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type BCSessionUnion struct {
	Bccs             *BCClientSession `protobuf:"bytes,1,opt" json:"Bccs,omitempty"`
	Bcss             *BCServerSession `protobuf:"bytes,2,opt" json:"Bcss,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *BCSessionUnion) Reset()         { *m = BCSessionUnion{} }
func (m *BCSessionUnion) String() string { return proto.CompactTextString(m) }
func (*BCSessionUnion) ProtoMessage()    {}

func (m *BCSessionUnion) GetBccs() *BCClientSession {
	if m != nil {
		return m.Bccs
	}
	return nil
}

func (m *BCSessionUnion) GetBcss() *BCServerSession {
	if m != nil {
		return m.Bcss
	}
	return nil
}

type BCClientSession struct {
	Dummy            *bool  `protobuf:"varint,1,opt" json:"Dummy,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *BCClientSession) Reset()         { *m = BCClientSession{} }
func (m *BCClientSession) String() string { return proto.CompactTextString(m) }
func (*BCClientSession) ProtoMessage()    {}

func (m *BCClientSession) GetDummy() bool {
	if m != nil && m.Dummy != nil {
		return *m.Dummy
	}
	return false
}

type BCServerSession struct {
	SArea            *int32 `protobuf:"varint,1,opt" json:"SArea,omitempty"`
	SType            *int32 `protobuf:"varint,2,opt" json:"SType,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *BCServerSession) Reset()         { *m = BCServerSession{} }
func (m *BCServerSession) String() string { return proto.CompactTextString(m) }
func (*BCServerSession) ProtoMessage()    {}

func (m *BCServerSession) GetSArea() int32 {
	if m != nil && m.SArea != nil {
		return *m.SArea
	}
	return 0
}

func (m *BCServerSession) GetSType() int32 {
	if m != nil && m.SType != nil {
		return *m.SType
	}
	return 0
}

func init() {
}
