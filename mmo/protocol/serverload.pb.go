// Code generated by protoc-gen-go.
// source: protocol/serverload.proto
// DO NOT EDIT!

package protocol

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type ServerLoad struct {
	SrvType          *int32 `protobuf:"varint,1,req" json:"SrvType,omitempty"`
	SrvId            *int32 `protobuf:"varint,2,req" json:"SrvId,omitempty"`
	CurLoad          *int32 `protobuf:"varint,3,req" json:"CurLoad,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ServerLoad) Reset()         { *m = ServerLoad{} }
func (m *ServerLoad) String() string { return proto.CompactTextString(m) }
func (*ServerLoad) ProtoMessage()    {}

func (m *ServerLoad) GetSrvType() int32 {
	if m != nil && m.SrvType != nil {
		return *m.SrvType
	}
	return 0
}

func (m *ServerLoad) GetSrvId() int32 {
	if m != nil && m.SrvId != nil {
		return *m.SrvId
	}
	return 0
}

func (m *ServerLoad) GetCurLoad() int32 {
	if m != nil && m.CurLoad != nil {
		return *m.CurLoad
	}
	return 0
}

type ServerStateSwitch struct {
	SrvType          *int32 `protobuf:"varint,1,req" json:"SrvType,omitempty"`
	SrvId            *int32 `protobuf:"varint,2,req" json:"SrvId,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ServerStateSwitch) Reset()         { *m = ServerStateSwitch{} }
func (m *ServerStateSwitch) String() string { return proto.CompactTextString(m) }
func (*ServerStateSwitch) ProtoMessage()    {}

func (m *ServerStateSwitch) GetSrvType() int32 {
	if m != nil && m.SrvType != nil {
		return *m.SrvType
	}
	return 0
}

func (m *ServerStateSwitch) GetSrvId() int32 {
	if m != nil && m.SrvId != nil {
		return *m.SrvId
	}
	return 0
}

func init() {
}