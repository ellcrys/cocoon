// Code generated by protoc-gen-go.
// source: server.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	server.proto

It has these top-level messages:
	LoginRequest
	AddCocoonToIdentityRequest
	CreateCocoonRequest
	GetCocoonRequest
	GetIdentityRequest
	CreateReleaseRequest
	GetReleaseRequest
	DeployRequest
	CreateIdentityRequest
	StopCocoonRequest
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

type LoginRequest struct {
	Email    string `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
}

func (m *LoginRequest) Reset()                    { *m = LoginRequest{} }
func (m *LoginRequest) String() string            { return proto1.CompactTextString(m) }
func (*LoginRequest) ProtoMessage()               {}
func (*LoginRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *LoginRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *LoginRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type AddCocoonToIdentityRequest struct {
	Email    string `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	CocoonId string `protobuf:"bytes,2,opt,name=cocoonId" json:"cocoonId,omitempty"`
}

func (m *AddCocoonToIdentityRequest) Reset()                    { *m = AddCocoonToIdentityRequest{} }
func (m *AddCocoonToIdentityRequest) String() string            { return proto1.CompactTextString(m) }
func (*AddCocoonToIdentityRequest) ProtoMessage()               {}
func (*AddCocoonToIdentityRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AddCocoonToIdentityRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *AddCocoonToIdentityRequest) GetCocoonId() string {
	if m != nil {
		return m.CocoonId
	}
	return ""
}

type CreateCocoonRequest struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	URL                  string   `protobuf:"bytes,2,opt,name=URL" json:"URL,omitempty"`
	Language             string   `protobuf:"bytes,3,opt,name=language" json:"language,omitempty"`
	ReleaseTag           string   `protobuf:"bytes,4,opt,name=releaseTag" json:"releaseTag,omitempty"`
	BuildParam           string   `protobuf:"bytes,5,opt,name=buildParam" json:"buildParam,omitempty"`
	Memory               string   `protobuf:"bytes,6,opt,name=memory" json:"memory,omitempty"`
	CPUShares            string   `protobuf:"bytes,7,opt,name=CPUShares" json:"CPUShares,omitempty"`
	Releases             []string `protobuf:"bytes,8,rep,name=releases" json:"releases,omitempty"`
	Link                 string   `protobuf:"bytes,9,opt,name=link" json:"link,omitempty"`
	NumSignatories       int32    `protobuf:"varint,10,opt,name=numSignatories" json:"numSignatories,omitempty"`
	SigThreshold         int32    `protobuf:"varint,11,opt,name=sigThreshold" json:"sigThreshold,omitempty"`
	Signatories          []string `protobuf:"bytes,12,rep,name=signatories" json:"signatories,omitempty"`
	Status               string   `protobuf:"bytes,13,opt,name=status" json:"status,omitempty"`
	CreatedAt            string   `protobuf:"bytes,14,opt,name=createdAt" json:"createdAt,omitempty"`
	OptionAllowDuplicate bool     `protobuf:"varint,15,opt,name=optionAllowDuplicate" json:"optionAllowDuplicate,omitempty"`
}

func (m *CreateCocoonRequest) Reset()                    { *m = CreateCocoonRequest{} }
func (m *CreateCocoonRequest) String() string            { return proto1.CompactTextString(m) }
func (*CreateCocoonRequest) ProtoMessage()               {}
func (*CreateCocoonRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *CreateCocoonRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *CreateCocoonRequest) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *CreateCocoonRequest) GetLanguage() string {
	if m != nil {
		return m.Language
	}
	return ""
}

func (m *CreateCocoonRequest) GetReleaseTag() string {
	if m != nil {
		return m.ReleaseTag
	}
	return ""
}

func (m *CreateCocoonRequest) GetBuildParam() string {
	if m != nil {
		return m.BuildParam
	}
	return ""
}

func (m *CreateCocoonRequest) GetMemory() string {
	if m != nil {
		return m.Memory
	}
	return ""
}

func (m *CreateCocoonRequest) GetCPUShares() string {
	if m != nil {
		return m.CPUShares
	}
	return ""
}

func (m *CreateCocoonRequest) GetReleases() []string {
	if m != nil {
		return m.Releases
	}
	return nil
}

func (m *CreateCocoonRequest) GetLink() string {
	if m != nil {
		return m.Link
	}
	return ""
}

func (m *CreateCocoonRequest) GetNumSignatories() int32 {
	if m != nil {
		return m.NumSignatories
	}
	return 0
}

func (m *CreateCocoonRequest) GetSigThreshold() int32 {
	if m != nil {
		return m.SigThreshold
	}
	return 0
}

func (m *CreateCocoonRequest) GetSignatories() []string {
	if m != nil {
		return m.Signatories
	}
	return nil
}

func (m *CreateCocoonRequest) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *CreateCocoonRequest) GetCreatedAt() string {
	if m != nil {
		return m.CreatedAt
	}
	return ""
}

func (m *CreateCocoonRequest) GetOptionAllowDuplicate() bool {
	if m != nil {
		return m.OptionAllowDuplicate
	}
	return false
}

type GetCocoonRequest struct {
	ID string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
}

func (m *GetCocoonRequest) Reset()                    { *m = GetCocoonRequest{} }
func (m *GetCocoonRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetCocoonRequest) ProtoMessage()               {}
func (*GetCocoonRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GetCocoonRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type GetIdentityRequest struct {
	Email string `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	ID    string `protobuf:"bytes,2,opt,name=ID" json:"ID,omitempty"`
}

func (m *GetIdentityRequest) Reset()                    { *m = GetIdentityRequest{} }
func (m *GetIdentityRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetIdentityRequest) ProtoMessage()               {}
func (*GetIdentityRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *GetIdentityRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *GetIdentityRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type CreateReleaseRequest struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	CocoonID             string   `protobuf:"bytes,2,opt,name=cocoonID" json:"cocoonID,omitempty"`
	URL                  string   `protobuf:"bytes,3,opt,name=URL" json:"URL,omitempty"`
	Language             string   `protobuf:"bytes,4,opt,name=language" json:"language,omitempty"`
	ReleaseTag           string   `protobuf:"bytes,5,opt,name=releaseTag" json:"releaseTag,omitempty"`
	BuildParam           string   `protobuf:"bytes,6,opt,name=buildParam" json:"buildParam,omitempty"`
	Link                 string   `protobuf:"bytes,7,opt,name=link" json:"link,omitempty"`
	SigApproved          int32    `protobuf:"varint,8,opt,name=sigApproved" json:"sigApproved,omitempty"`
	SigDenied            int32    `protobuf:"varint,9,opt,name=sigDenied" json:"sigDenied,omitempty"`
	CreatedAt            string   `protobuf:"bytes,10,opt,name=createdAt" json:"createdAt,omitempty"`
	VotersID             []string `protobuf:"bytes,11,rep,name=votersID" json:"votersID,omitempty"`
	OptionAllowDuplicate bool     `protobuf:"varint,12,opt,name=optionAllowDuplicate" json:"optionAllowDuplicate,omitempty"`
}

func (m *CreateReleaseRequest) Reset()                    { *m = CreateReleaseRequest{} }
func (m *CreateReleaseRequest) String() string            { return proto1.CompactTextString(m) }
func (*CreateReleaseRequest) ProtoMessage()               {}
func (*CreateReleaseRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *CreateReleaseRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *CreateReleaseRequest) GetCocoonID() string {
	if m != nil {
		return m.CocoonID
	}
	return ""
}

func (m *CreateReleaseRequest) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *CreateReleaseRequest) GetLanguage() string {
	if m != nil {
		return m.Language
	}
	return ""
}

func (m *CreateReleaseRequest) GetReleaseTag() string {
	if m != nil {
		return m.ReleaseTag
	}
	return ""
}

func (m *CreateReleaseRequest) GetBuildParam() string {
	if m != nil {
		return m.BuildParam
	}
	return ""
}

func (m *CreateReleaseRequest) GetLink() string {
	if m != nil {
		return m.Link
	}
	return ""
}

func (m *CreateReleaseRequest) GetSigApproved() int32 {
	if m != nil {
		return m.SigApproved
	}
	return 0
}

func (m *CreateReleaseRequest) GetSigDenied() int32 {
	if m != nil {
		return m.SigDenied
	}
	return 0
}

func (m *CreateReleaseRequest) GetCreatedAt() string {
	if m != nil {
		return m.CreatedAt
	}
	return ""
}

func (m *CreateReleaseRequest) GetVotersID() []string {
	if m != nil {
		return m.VotersID
	}
	return nil
}

func (m *CreateReleaseRequest) GetOptionAllowDuplicate() bool {
	if m != nil {
		return m.OptionAllowDuplicate
	}
	return false
}

type GetReleaseRequest struct {
	ID string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
}

func (m *GetReleaseRequest) Reset()                    { *m = GetReleaseRequest{} }
func (m *GetReleaseRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetReleaseRequest) ProtoMessage()               {}
func (*GetReleaseRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetReleaseRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type DeployRequest struct {
	CocoonID   string `protobuf:"bytes,1,opt,name=cocoonID" json:"cocoonID,omitempty"`
	URL        string `protobuf:"bytes,2,opt,name=URL" json:"URL,omitempty"`
	Language   string `protobuf:"bytes,3,opt,name=language" json:"language,omitempty"`
	ReleaseTag string `protobuf:"bytes,4,opt,name=releaseTag" json:"releaseTag,omitempty"`
	BuildParam []byte `protobuf:"bytes,5,opt,name=buildParam,proto3" json:"buildParam,omitempty"`
	Memory     string `protobuf:"bytes,6,opt,name=memory" json:"memory,omitempty"`
	CPUShares  string `protobuf:"bytes,7,opt,name=CPUShares" json:"CPUShares,omitempty"`
	Link       string `protobuf:"bytes,8,opt,name=link" json:"link,omitempty"`
}

func (m *DeployRequest) Reset()                    { *m = DeployRequest{} }
func (m *DeployRequest) String() string            { return proto1.CompactTextString(m) }
func (*DeployRequest) ProtoMessage()               {}
func (*DeployRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *DeployRequest) GetCocoonID() string {
	if m != nil {
		return m.CocoonID
	}
	return ""
}

func (m *DeployRequest) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *DeployRequest) GetLanguage() string {
	if m != nil {
		return m.Language
	}
	return ""
}

func (m *DeployRequest) GetReleaseTag() string {
	if m != nil {
		return m.ReleaseTag
	}
	return ""
}

func (m *DeployRequest) GetBuildParam() []byte {
	if m != nil {
		return m.BuildParam
	}
	return nil
}

func (m *DeployRequest) GetMemory() string {
	if m != nil {
		return m.Memory
	}
	return ""
}

func (m *DeployRequest) GetCPUShares() string {
	if m != nil {
		return m.CPUShares
	}
	return ""
}

func (m *DeployRequest) GetLink() string {
	if m != nil {
		return m.Link
	}
	return ""
}

type CreateIdentityRequest struct {
	Email                string   `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	Password             string   `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
	Cocoons              []string `protobuf:"bytes,3,rep,name=cocoons" json:"cocoons,omitempty"`
	ClientSessions       []string `protobuf:"bytes,4,rep,name=clientSessions" json:"clientSessions,omitempty"`
	OptionAllowDuplicate bool     `protobuf:"varint,5,opt,name=optionAllowDuplicate" json:"optionAllowDuplicate,omitempty"`
}

func (m *CreateIdentityRequest) Reset()                    { *m = CreateIdentityRequest{} }
func (m *CreateIdentityRequest) String() string            { return proto1.CompactTextString(m) }
func (*CreateIdentityRequest) ProtoMessage()               {}
func (*CreateIdentityRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *CreateIdentityRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *CreateIdentityRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *CreateIdentityRequest) GetCocoons() []string {
	if m != nil {
		return m.Cocoons
	}
	return nil
}

func (m *CreateIdentityRequest) GetClientSessions() []string {
	if m != nil {
		return m.ClientSessions
	}
	return nil
}

func (m *CreateIdentityRequest) GetOptionAllowDuplicate() bool {
	if m != nil {
		return m.OptionAllowDuplicate
	}
	return false
}

type StopCocoonRequest struct {
	ID string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
}

func (m *StopCocoonRequest) Reset()                    { *m = StopCocoonRequest{} }
func (m *StopCocoonRequest) String() string            { return proto1.CompactTextString(m) }
func (*StopCocoonRequest) ProtoMessage()               {}
func (*StopCocoonRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *StopCocoonRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type Response struct {
	Status int32  `protobuf:"varint,1,opt,name=status" json:"status,omitempty"`
	Body   []byte `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto1.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *Response) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *Response) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto1.RegisterType((*LoginRequest)(nil), "proto.LoginRequest")
	proto1.RegisterType((*AddCocoonToIdentityRequest)(nil), "proto.AddCocoonToIdentityRequest")
	proto1.RegisterType((*CreateCocoonRequest)(nil), "proto.CreateCocoonRequest")
	proto1.RegisterType((*GetCocoonRequest)(nil), "proto.GetCocoonRequest")
	proto1.RegisterType((*GetIdentityRequest)(nil), "proto.GetIdentityRequest")
	proto1.RegisterType((*CreateReleaseRequest)(nil), "proto.CreateReleaseRequest")
	proto1.RegisterType((*GetReleaseRequest)(nil), "proto.GetReleaseRequest")
	proto1.RegisterType((*DeployRequest)(nil), "proto.DeployRequest")
	proto1.RegisterType((*CreateIdentityRequest)(nil), "proto.CreateIdentityRequest")
	proto1.RegisterType((*StopCocoonRequest)(nil), "proto.StopCocoonRequest")
	proto1.RegisterType((*Response)(nil), "proto.Response")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for API service

type APIClient interface {
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*Response, error)
	CreateCocoon(ctx context.Context, in *CreateCocoonRequest, opts ...grpc.CallOption) (*Response, error)
	CreateRelease(ctx context.Context, in *CreateReleaseRequest, opts ...grpc.CallOption) (*Response, error)
	GetRelease(ctx context.Context, in *GetReleaseRequest, opts ...grpc.CallOption) (*Response, error)
	CreateIdentity(ctx context.Context, in *CreateIdentityRequest, opts ...grpc.CallOption) (*Response, error)
	Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*Response, error)
	GetCocoon(ctx context.Context, in *GetCocoonRequest, opts ...grpc.CallOption) (*Response, error)
	GetIdentity(ctx context.Context, in *GetIdentityRequest, opts ...grpc.CallOption) (*Response, error)
	AddCocoonToIdentity(ctx context.Context, in *AddCocoonToIdentityRequest, opts ...grpc.CallOption) (*Response, error)
	StopCocoon(ctx context.Context, in *StopCocoonRequest, opts ...grpc.CallOption) (*Response, error)
}

type aPIClient struct {
	cc *grpc.ClientConn
}

func NewAPIClient(cc *grpc.ClientConn) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/Login", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) CreateCocoon(ctx context.Context, in *CreateCocoonRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/CreateCocoon", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) CreateRelease(ctx context.Context, in *CreateReleaseRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/CreateRelease", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) GetRelease(ctx context.Context, in *GetReleaseRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/GetRelease", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) CreateIdentity(ctx context.Context, in *CreateIdentityRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/CreateIdentity", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/Deploy", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) GetCocoon(ctx context.Context, in *GetCocoonRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/GetCocoon", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) GetIdentity(ctx context.Context, in *GetIdentityRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/GetIdentity", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) AddCocoonToIdentity(ctx context.Context, in *AddCocoonToIdentityRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/AddCocoonToIdentity", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) StopCocoon(ctx context.Context, in *StopCocoonRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.API/StopCocoon", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for API service

type APIServer interface {
	Login(context.Context, *LoginRequest) (*Response, error)
	CreateCocoon(context.Context, *CreateCocoonRequest) (*Response, error)
	CreateRelease(context.Context, *CreateReleaseRequest) (*Response, error)
	GetRelease(context.Context, *GetReleaseRequest) (*Response, error)
	CreateIdentity(context.Context, *CreateIdentityRequest) (*Response, error)
	Deploy(context.Context, *DeployRequest) (*Response, error)
	GetCocoon(context.Context, *GetCocoonRequest) (*Response, error)
	GetIdentity(context.Context, *GetIdentityRequest) (*Response, error)
	AddCocoonToIdentity(context.Context, *AddCocoonToIdentityRequest) (*Response, error)
	StopCocoon(context.Context, *StopCocoonRequest) (*Response, error)
}

func RegisterAPIServer(s *grpc.Server, srv APIServer) {
	s.RegisterService(&_API_serviceDesc, srv)
}

func _API_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_CreateCocoon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCocoonRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).CreateCocoon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/CreateCocoon",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).CreateCocoon(ctx, req.(*CreateCocoonRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_CreateRelease_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).CreateRelease(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/CreateRelease",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).CreateRelease(ctx, req.(*CreateReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_GetRelease_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetRelease(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/GetRelease",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetRelease(ctx, req.(*GetReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_CreateIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateIdentityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).CreateIdentity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/CreateIdentity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).CreateIdentity(ctx, req.(*CreateIdentityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_Deploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).Deploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/Deploy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).Deploy(ctx, req.(*DeployRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_GetCocoon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCocoonRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetCocoon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/GetCocoon",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetCocoon(ctx, req.(*GetCocoonRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_GetIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIdentityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetIdentity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/GetIdentity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetIdentity(ctx, req.(*GetIdentityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_AddCocoonToIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddCocoonToIdentityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).AddCocoonToIdentity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/AddCocoonToIdentity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).AddCocoonToIdentity(ctx, req.(*AddCocoonToIdentityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_StopCocoon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopCocoonRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).StopCocoon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.API/StopCocoon",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).StopCocoon(ctx, req.(*StopCocoonRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _API_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.API",
	HandlerType: (*APIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _API_Login_Handler,
		},
		{
			MethodName: "CreateCocoon",
			Handler:    _API_CreateCocoon_Handler,
		},
		{
			MethodName: "CreateRelease",
			Handler:    _API_CreateRelease_Handler,
		},
		{
			MethodName: "GetRelease",
			Handler:    _API_GetRelease_Handler,
		},
		{
			MethodName: "CreateIdentity",
			Handler:    _API_CreateIdentity_Handler,
		},
		{
			MethodName: "Deploy",
			Handler:    _API_Deploy_Handler,
		},
		{
			MethodName: "GetCocoon",
			Handler:    _API_GetCocoon_Handler,
		},
		{
			MethodName: "GetIdentity",
			Handler:    _API_GetIdentity_Handler,
		},
		{
			MethodName: "AddCocoonToIdentity",
			Handler:    _API_AddCocoonToIdentity_Handler,
		},
		{
			MethodName: "StopCocoon",
			Handler:    _API_StopCocoon_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}

func init() { proto1.RegisterFile("server.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 746 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x56, 0xdd, 0x6e, 0xd3, 0x4a,
	0x10, 0x96, 0x93, 0x38, 0x4d, 0x26, 0x6e, 0xda, 0xb3, 0xcd, 0x39, 0x67, 0x4f, 0x4e, 0x85, 0x82,
	0x91, 0x50, 0x6f, 0x28, 0x52, 0x11, 0x20, 0x40, 0x08, 0xa2, 0x46, 0xaa, 0x22, 0x55, 0xa8, 0x72,
	0xda, 0x07, 0x70, 0xe3, 0x51, 0xba, 0xc2, 0xf1, 0x1a, 0xef, 0xa6, 0x55, 0x9f, 0x8e, 0x67, 0xe8,
	0x33, 0xf0, 0x06, 0x3c, 0x01, 0xf2, 0xae, 0x7f, 0xe3, 0xfc, 0xf4, 0x02, 0xae, 0xe2, 0x99, 0xd9,
	0x6f, 0x76, 0x77, 0xbe, 0x6f, 0x67, 0x02, 0x96, 0xc0, 0xe8, 0x16, 0xa3, 0xe3, 0x30, 0xe2, 0x92,
	0x13, 0x53, 0xfd, 0xd8, 0x9f, 0xc1, 0x3a, 0xe7, 0x33, 0x16, 0x38, 0xf8, 0x6d, 0x81, 0x42, 0x92,
	0x1e, 0x98, 0x38, 0x77, 0x99, 0x4f, 0x8d, 0x81, 0x71, 0xd4, 0x76, 0xb4, 0x41, 0xfa, 0xd0, 0x0a,
	0x5d, 0x21, 0xee, 0x78, 0xe4, 0xd1, 0x9a, 0x0a, 0x64, 0xb6, 0xfd, 0x05, 0xfa, 0x43, 0xcf, 0x3b,
	0xe5, 0x53, 0xce, 0x83, 0x4b, 0x3e, 0xf6, 0x30, 0x90, 0x4c, 0xde, 0x6f, 0xcd, 0x37, 0x55, 0x80,
	0x71, 0x96, 0x2f, 0xb5, 0xed, 0x87, 0x3a, 0x1c, 0x9c, 0x46, 0xe8, 0x4a, 0xd4, 0x39, 0xd3, 0x4c,
	0x5d, 0xa8, 0x8d, 0x47, 0x49, 0x9a, 0xda, 0x78, 0x44, 0xf6, 0xa1, 0x7e, 0xe5, 0x9c, 0x27, 0xf0,
	0xf8, 0x33, 0xce, 0xea, 0xbb, 0xc1, 0x6c, 0xe1, 0xce, 0x90, 0xd6, 0x75, 0xd6, 0xd4, 0x26, 0x4f,
	0x00, 0x22, 0xf4, 0xd1, 0x15, 0x78, 0xe9, 0xce, 0x68, 0x43, 0x45, 0x0b, 0x9e, 0x38, 0x7e, 0xbd,
	0x60, 0xbe, 0x77, 0xe1, 0x46, 0xee, 0x9c, 0x9a, 0x3a, 0x9e, 0x7b, 0xc8, 0x3f, 0xd0, 0x9c, 0xe3,
	0x9c, 0x47, 0xf7, 0xb4, 0xa9, 0x62, 0x89, 0x45, 0x0e, 0xa1, 0x7d, 0x7a, 0x71, 0x35, 0xb9, 0x71,
	0x23, 0x14, 0x74, 0x47, 0x85, 0x72, 0x47, 0x7c, 0xa2, 0x64, 0x0f, 0x41, 0x5b, 0x83, 0x7a, 0x7c,
	0xa2, 0xd4, 0x26, 0x04, 0x1a, 0x3e, 0x0b, 0xbe, 0xd2, 0xb6, 0x02, 0xa9, 0x6f, 0xf2, 0x1c, 0xba,
	0xc1, 0x62, 0x3e, 0x61, 0xb3, 0xc0, 0x95, 0x3c, 0x62, 0x28, 0x28, 0x0c, 0x8c, 0x23, 0xd3, 0x59,
	0xf2, 0x12, 0x1b, 0x2c, 0xc1, 0x66, 0x97, 0x37, 0x11, 0x8a, 0x1b, 0xee, 0x7b, 0xb4, 0xa3, 0x56,
	0x95, 0x7c, 0x64, 0x00, 0x1d, 0x51, 0x48, 0x64, 0xa9, 0xed, 0x8b, 0xae, 0xf8, 0x4e, 0x42, 0xba,
	0x72, 0x21, 0xe8, 0xae, 0xbe, 0x93, 0xb6, 0xe2, 0x3b, 0x4d, 0x15, 0x01, 0xde, 0x50, 0xd2, 0xae,
	0xbe, 0x53, 0xe6, 0x20, 0x27, 0xd0, 0xe3, 0xa1, 0x64, 0x3c, 0x18, 0xfa, 0x3e, 0xbf, 0x1b, 0x2d,
	0x42, 0x9f, 0x4d, 0x5d, 0x89, 0x74, 0x6f, 0x60, 0x1c, 0xb5, 0x9c, 0x95, 0x31, 0xdb, 0x86, 0xfd,
	0x33, 0x94, 0x1b, 0xf9, 0xb4, 0xdf, 0x03, 0x39, 0x43, 0xf9, 0x38, 0xfd, 0x68, 0x6c, 0x2d, 0xc3,
	0xfe, 0xac, 0x41, 0x4f, 0x6b, 0xc6, 0xd1, 0xe5, 0x5d, 0x27, 0x9a, 0x5c, 0x78, 0xa3, 0x25, 0xe1,
	0x65, 0x82, 0xaa, 0xaf, 0x16, 0x54, 0x63, 0xa3, 0xa0, 0xcc, 0x2d, 0x82, 0x6a, 0x56, 0x04, 0x95,
	0xd2, 0xbf, 0x53, 0xa0, 0x5f, 0x53, 0x36, 0x0c, 0xc3, 0x88, 0xdf, 0xa2, 0x47, 0x5b, 0x8a, 0xd5,
	0xa2, 0x2b, 0xa6, 0x46, 0xb0, 0xd9, 0x08, 0x03, 0x86, 0x9e, 0x52, 0x8e, 0xe9, 0xe4, 0x8e, 0x32,
	0x71, 0xb0, 0x4c, 0x5c, 0x1f, 0x5a, 0xb7, 0x5c, 0x62, 0x24, 0xc6, 0x23, 0xda, 0xd1, 0x62, 0x4c,
	0xed, 0xb5, 0xa4, 0x5a, 0x1b, 0x48, 0x7d, 0x06, 0x7f, 0x9d, 0xa1, 0xdc, 0x5c, 0x70, 0xfb, 0x87,
	0x01, 0xbb, 0x23, 0x0c, 0x7d, 0x9e, 0x31, 0x5a, 0xa4, 0xc0, 0x58, 0x4d, 0xc1, 0x1f, 0x7b, 0xd3,
	0xd6, 0x6f, 0x78, 0xd3, 0x29, 0x71, 0xad, 0x9c, 0x38, 0xfb, 0xbb, 0x01, 0x7f, 0x6b, 0xfd, 0x3d,
	0xba, 0xff, 0xad, 0xeb, 0xa7, 0x84, 0xc2, 0x8e, 0xae, 0x87, 0xa0, 0x75, 0xc5, 0x52, 0x6a, 0xc6,
	0xdd, 0x61, 0xea, 0x33, 0x0c, 0xe4, 0x04, 0x85, 0x60, 0xf1, 0x82, 0x86, 0x5a, 0xb0, 0xe4, 0x5d,
	0x4b, 0xa6, 0xb9, 0x99, 0xcc, 0x89, 0xe4, 0xe1, 0xe6, 0x27, 0xfa, 0x06, 0x5a, 0x0e, 0x8a, 0x90,
	0x07, 0x02, 0x0b, 0xcd, 0xc3, 0x50, 0x32, 0x4c, 0x9b, 0x07, 0x81, 0xc6, 0x35, 0xf7, 0xee, 0xd5,
	0xb5, 0x2c, 0x47, 0x7d, 0x9f, 0x3c, 0x34, 0xa0, 0x3e, 0xbc, 0x18, 0x93, 0x17, 0x60, 0xaa, 0x61,
	0x43, 0x0e, 0xf4, 0x10, 0x3a, 0x2e, 0x8e, 0x9e, 0xfe, 0x5e, 0xe2, 0xcc, 0xb6, 0xf8, 0x00, 0x56,
	0x71, 0x10, 0x90, 0x7e, 0xb2, 0x60, 0xc5, 0x74, 0xa8, 0x82, 0x3f, 0xc2, 0x6e, 0xa9, 0x23, 0x90,
	0xff, 0x4b, 0xe8, 0xb2, 0x6c, 0xab, 0xf0, 0xb7, 0x00, 0xb9, 0xb8, 0x09, 0x4d, 0xc2, 0x15, 0xbd,
	0x57, 0x81, 0x9f, 0xa0, 0x5b, 0x56, 0x02, 0x39, 0x2c, 0x6d, 0xbc, 0x24, 0x90, 0x6a, 0x82, 0x97,
	0xd0, 0xd4, 0x0f, 0x86, 0xf4, 0x92, 0x50, 0xe9, 0xfd, 0x54, 0x01, 0xaf, 0xa1, 0x9d, 0x35, 0x57,
	0xf2, 0x6f, 0x7e, 0xd2, 0x2d, 0x05, 0x7a, 0x07, 0x9d, 0x42, 0xbf, 0x25, 0xff, 0xe5, 0xc0, 0xad,
	0x47, 0x1c, 0xc3, 0xc1, 0x8a, 0x91, 0x4f, 0x9e, 0x26, 0xeb, 0xd6, 0xff, 0x1d, 0x58, 0x59, 0xe7,
	0x5c, 0x77, 0x59, 0x9d, 0x2b, 0x52, 0xac, 0x00, 0xaf, 0x9b, 0xca, 0x7e, 0xf5, 0x2b, 0x00, 0x00,
	0xff, 0xff, 0xb9, 0x44, 0x5c, 0x5b, 0xd6, 0x08, 0x00, 0x00,
}
