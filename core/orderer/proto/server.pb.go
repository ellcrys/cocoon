// Code generated by protoc-gen-go.
// source: server.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	server.proto

It has these top-level messages:
	OrdererTx
	Response
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type OrdererTx struct {
	Id    string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Value string `protobuf:"bytes,3,opt,name=value" json:"value,omitempty"`
}

func (m *OrdererTx) Reset()                    { *m = OrdererTx{} }
func (m *OrdererTx) String() string            { return proto1.CompactTextString(m) }
func (*OrdererTx) ProtoMessage()               {}
func (*OrdererTx) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *OrdererTx) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *OrdererTx) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *OrdererTx) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Response struct {
	Code    int32  `protobuf:"varint,1,opt,name=code" json:"code,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	Data    []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto1.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Response) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *Response) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Response) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto1.RegisterType((*OrdererTx)(nil), "proto.OrdererTx")
	proto1.RegisterType((*Response)(nil), "proto.Response")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Orderer service

type OrdererClient interface {
	Put(ctx context.Context, in *OrdererTx, opts ...grpc.CallOption) (*Response, error)
}

type ordererClient struct {
	cc *grpc.ClientConn
}

func NewOrdererClient(cc *grpc.ClientConn) OrdererClient {
	return &ordererClient{cc}
}

func (c *ordererClient) Put(ctx context.Context, in *OrdererTx, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.Orderer/Put", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Orderer service

type OrdererServer interface {
	Put(context.Context, *OrdererTx) (*Response, error)
}

func RegisterOrdererServer(s *grpc.Server, srv OrdererServer) {
	s.RegisterService(&_Orderer_serviceDesc, srv)
}

func _Orderer_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OrdererTx)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdererServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Orderer/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdererServer).Put(ctx, req.(*OrdererTx))
	}
	return interceptor(ctx, in, info, handler)
}

var _Orderer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Orderer",
	HandlerType: (*OrdererServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Put",
			Handler:    _Orderer_Put_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}

func init() { proto1.RegisterFile("server.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 178 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x3c, 0x8e, 0x41, 0x0b, 0x82, 0x40,
	0x10, 0x85, 0x51, 0x33, 0x73, 0x90, 0x8a, 0xa1, 0xc3, 0xd2, 0x29, 0x3c, 0x79, 0xf2, 0x90, 0xbf,
	0xa1, 0x5b, 0x50, 0x2c, 0xfd, 0x81, 0xad, 0x1d, 0x42, 0x48, 0x57, 0x76, 0x55, 0xfa, 0xf9, 0xe1,
	0xe8, 0x76, 0x9a, 0xf7, 0x1e, 0x33, 0x6f, 0x3e, 0xc8, 0x1c, 0xd9, 0x91, 0x6c, 0xd9, 0x59, 0xd3,
	0x1b, 0x8c, 0x79, 0xe4, 0x17, 0x48, 0x6f, 0x56, 0x93, 0x25, 0xfb, 0xf8, 0xe2, 0x16, 0xc2, 0x5a,
	0x8b, 0xe0, 0x14, 0x14, 0xa9, 0x0c, 0x6b, 0x8d, 0x08, 0xab, 0x56, 0x35, 0x24, 0x42, 0x4e, 0x58,
	0xe3, 0x01, 0xe2, 0x51, 0x7d, 0x06, 0x12, 0x11, 0x87, 0xb3, 0xc9, 0xaf, 0xb0, 0x91, 0xe4, 0x3a,
	0xd3, 0x3a, 0x9a, 0xae, 0x5e, 0x46, 0x13, 0xf7, 0xc4, 0x92, 0x35, 0x0a, 0x48, 0x1a, 0x72, 0x4e,
	0xbd, 0x7d, 0x99, 0xb7, 0xd3, 0xb6, 0x56, 0xbd, 0xe2, 0xba, 0x4c, 0xb2, 0x3e, 0x57, 0x90, 0x2c,
	0x50, 0x58, 0x40, 0x74, 0x1f, 0x7a, 0xdc, 0xcf, 0xd4, 0xe5, 0x9f, 0xf5, 0xb8, 0x5b, 0x12, 0xff,
	0xf6, 0xb9, 0x66, 0x5f, 0xfd, 0x02, 0x00, 0x00, 0xff, 0xff, 0x37, 0xdd, 0x04, 0xdc, 0xe7, 0x00,
	0x00, 0x00,
}