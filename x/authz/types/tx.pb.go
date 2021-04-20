package types

import (
	context "context"
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	types1 "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "github.com/regen-network/cosmos-proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// MsgGrantRequest grants the provided authorization to the grantee on the granter's
// account with the provided expiration time.
type MsgGrantRequest struct {
	Granter       string     `protobuf:"bytes,1,opt,name=granter,proto3" json:"granter,omitempty"`
	Grantee       string     `protobuf:"bytes,2,opt,name=grantee,proto3" json:"grantee,omitempty"`
	Authorization *types.Any `protobuf:"bytes,3,opt,name=authorization,proto3" json:"authorization,omitempty"`
	Expiration    time.Time  `protobuf:"bytes,4,opt,name=expiration,proto3,stdtime" json:"expiration"`
}

func (m *MsgGrantRequest) Reset()         { *m = MsgGrantRequest{} }
func (m *MsgGrantRequest) String() string { return proto.CompactTextString(m) }
func (*MsgGrantRequest) ProtoMessage()    {}
func (*MsgGrantRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3ceddab7d8589ad1, []int{0}
}
func (m *MsgGrantRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgGrantRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgGrantRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgGrantRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgGrantRequest.Merge(m, src)
}
func (m *MsgGrantRequest) XXX_Size() int {
	return m.Size()
}
func (m *MsgGrantRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgGrantRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MsgGrantRequest proto.InternalMessageInfo

func (m *MsgGrantRequest) GetGranter() string {
	if m != nil {
		return m.Granter
	}
	return ""
}

func (m *MsgGrantRequest) GetGrantee() string {
	if m != nil {
		return m.Grantee
	}
	return ""
}

func (m *MsgGrantRequest) GetAuthorization() *types.Any {
	if m != nil {
		return m.Authorization
	}
	return nil
}

func (m *MsgGrantRequest) GetExpiration() time.Time {
	if m != nil {
		return m.Expiration
	}
	return time.Time{}
}

// MsgExecResponse defines the Msg/MsgExecResponse response type.
type MsgExecResponse struct {
	Result *types1.Result `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (m *MsgExecResponse) Reset()         { *m = MsgExecResponse{} }
func (m *MsgExecResponse) String() string { return proto.CompactTextString(m) }
func (*MsgExecResponse) ProtoMessage()    {}
func (*MsgExecResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3ceddab7d8589ad1, []int{1}
}
func (m *MsgExecResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgExecResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgExecAuthorizedResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgExecResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgExecAuthorizedResponse.Merge(m, src)
}
func (m *MsgExecResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgExecResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgExecAuthorizedResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgExecAuthorizedResponse proto.InternalMessageInfo

func (m *MsgExecResponse) GetResult() *types1.Result {
	if m != nil {
		return m.Result
	}
	return nil
}

// MsgExecRequest attempts to execute the provided messages using
// authorizations granted to the grantee. Each message should have only
// one signer corresponding to the granter of the authorization.
type MsgExecRequest struct {
	Grantee string       `protobuf:"bytes,1,opt,name=grantee,proto3" json:"grantee,omitempty"`
	Msgs    []*types.Any `protobuf:"bytes,2,rep,name=msgs,proto3" json:"msgs,omitempty"`
}

func (m *MsgExecRequest) Reset()         { *m = MsgExecRequest{} }
func (m *MsgExecRequest) String() string { return proto.CompactTextString(m) }
func (*MsgExecRequest) ProtoMessage()    {}
func (*MsgExecRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3ceddab7d8589ad1, []int{2}
}
func (m *MsgExecRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgExecRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgExecAuthorizedRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgExecRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgExecAuthorizedRequest.Merge(m, src)
}
func (m *MsgExecRequest) XXX_Size() int {
	return m.Size()
}
func (m *MsgExecRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgExecAuthorizedRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MsgExecAuthorizedRequest proto.InternalMessageInfo

func (m *MsgExecRequest) GetGrantee() string {
	if m != nil {
		return m.Grantee
	}
	return ""
}

func (m *MsgExecRequest) GetMsgs() []*types.Any {
	if m != nil {
		return m.Msgs
	}
	return nil
}

// MsgGrantResponse defines the Msg/MsgGrant response type.
type MsgGrantResponse struct {
}

func (m *MsgGrantResponse) Reset()         { *m = MsgGrantResponse{} }
func (m *MsgGrantResponse) String() string { return proto.CompactTextString(m) }
func (*MsgGrantResponse) ProtoMessage()    {}
func (*MsgGrantResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3ceddab7d8589ad1, []int{3}
}
func (m *MsgGrantResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgGrantResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgGrantResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgGrantResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgGrantResponse.Merge(m, src)
}
func (m *MsgGrantResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgGrantResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgGrantResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgGrantResponse proto.InternalMessageInfo

// MsgRevokeRequest revokes any authorization with the provided sdk.Msg type on the
// granter's account with that has been granted to the grantee.
type MsgRevokeRequest struct {
	Granter    string `protobuf:"bytes,1,opt,name=granter,proto3" json:"granter,omitempty"`
	Grantee    string `protobuf:"bytes,2,opt,name=grantee,proto3" json:"grantee,omitempty"`
	MsgTypeUrl string `protobuf:"bytes,3,opt,name=method_name,json=methodName,proto3" json:"method_name,omitempty"`
}

func (m *MsgRevokeRequest) Reset()         { *m = MsgRevokeRequest{} }
func (m *MsgRevokeRequest) String() string { return proto.CompactTextString(m) }
func (*MsgRevokeRequest) ProtoMessage()    {}
func (*MsgRevokeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3ceddab7d8589ad1, []int{4}
}
func (m *MsgRevokeRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRevokeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRevokeRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRevokeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRevokeRequest.Merge(m, src)
}
func (m *MsgRevokeRequest) XXX_Size() int {
	return m.Size()
}
func (m *MsgRevokeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRevokeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRevokeRequest proto.InternalMessageInfo

func (m *MsgRevokeRequest) GetGranter() string {
	if m != nil {
		return m.Granter
	}
	return ""
}

func (m *MsgRevokeRequest) GetGrantee() string {
	if m != nil {
		return m.Grantee
	}
	return ""
}

func (m *MsgRevokeRequest) GetMethodName() string {
	if m != nil {
		return m.MsgTypeUrl
	}
	return ""
}

// MsgRevokeResponse defines the Msg/MsgRevokeResponse response type.
type MsgRevokeResponse struct {
}

func (m *MsgRevokeResponse) Reset()         { *m = MsgRevokeResponse{} }
func (m *MsgRevokeResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRevokeResponse) ProtoMessage()    {}
func (*MsgRevokeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3ceddab7d8589ad1, []int{5}
}
func (m *MsgRevokeResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRevokeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRevokeResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRevokeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRevokeResponse.Merge(m, src)
}
func (m *MsgRevokeResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRevokeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRevokeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRevokeResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgGrantRequest)(nil), "cosmos.authz.v1beta1.MsgGrantRequest")
	proto.RegisterType((*MsgExecResponse)(nil), "cosmos.authz.v1beta1.MsgExecAuthorizedResponse")
	proto.RegisterType((*MsgExecRequest)(nil), "cosmos.authz.v1beta1.MsgExecAuthorizedRequest")
	proto.RegisterType((*MsgGrantResponse)(nil), "cosmos.authz.v1beta1.MsgGrantResponse")
	proto.RegisterType((*MsgRevokeRequest)(nil), "cosmos.authz.v1beta1.MsgRevokeRequest")
	proto.RegisterType((*MsgRevokeResponse)(nil), "cosmos.authz.v1beta1.MsgRevokeResponse")
}

func init() { proto.RegisterFile("cosmos/authz/v1beta1/tx.proto", fileDescriptor_3ceddab7d8589ad1) }

var fileDescriptor_3ceddab7d8589ad1 = []byte{
	// 514 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x8d, 0x93, 0x12, 0xe8, 0x46, 0x05, 0xba, 0xe4, 0xe0, 0x58, 0xc2, 0x89, 0x8c, 0x80, 0x08,
	0xa9, 0x6b, 0x35, 0x5c, 0xb8, 0x36, 0x2a, 0xe2, 0x14, 0x0e, 0x2b, 0x40, 0x82, 0x03, 0xd5, 0x3a,
	0x19, 0x36, 0x56, 0x6b, 0x6f, 0xea, 0x5d, 0x57, 0x49, 0xbf, 0xa2, 0xff, 0xc1, 0x95, 0x8f, 0xa8,
	0x38, 0xf5, 0x88, 0x38, 0x00, 0x4a, 0x7e, 0x04, 0x79, 0x77, 0x43, 0x93, 0x90, 0xa8, 0x48, 0x9c,
	0xec, 0x99, 0xf7, 0x76, 0xe6, 0xed, 0x9b, 0xb1, 0xd1, 0xc3, 0xbe, 0x90, 0x89, 0x90, 0x21, 0xcb,
	0xd5, 0xf0, 0x3c, 0x3c, 0xdb, 0x8f, 0x40, 0xb1, 0xfd, 0x50, 0x8d, 0xc9, 0x28, 0x13, 0x4a, 0xe0,
	0xba, 0x81, 0x89, 0x86, 0x89, 0x85, 0xbd, 0x86, 0xc9, 0x1e, 0x69, 0x4e, 0x68, 0x29, 0x3a, 0xf0,
	0xea, 0x5c, 0x70, 0x61, 0xf2, 0xc5, 0x9b, 0xcd, 0x36, 0xb9, 0x10, 0xfc, 0x04, 0x42, 0x1d, 0x45,
	0xf9, 0xa7, 0x50, 0xc5, 0x09, 0x48, 0xc5, 0x92, 0x91, 0x25, 0x34, 0x56, 0x09, 0x2c, 0x9d, 0x58,
	0xe8, 0x91, 0x55, 0x18, 0x31, 0x09, 0x21, 0x8b, 0xfa, 0xf1, 0x1f, 0x95, 0x45, 0x60, 0x48, 0xc1,
	0x77, 0x07, 0xdd, 0xeb, 0x49, 0xfe, 0x2a, 0x63, 0xa9, 0xa2, 0x70, 0x9a, 0x83, 0x54, 0xd8, 0x45,
	0xb7, 0x79, 0x11, 0x43, 0xe6, 0x3a, 0x2d, 0xa7, 0xbd, 0x4d, 0xe7, 0xe1, 0x35, 0x02, 0x6e, 0x79,
	0x11, 0x01, 0xdc, 0x43, 0x3b, 0xc5, 0x55, 0x45, 0x16, 0x9f, 0x33, 0x15, 0x8b, 0xd4, 0xad, 0xb4,
	0x9c, 0x76, 0xad, 0x53, 0x27, 0x46, 0x1f, 0x99, 0xeb, 0x23, 0x07, 0xe9, 0xa4, 0xbb, 0xfb, 0xf5,
	0xcb, 0xde, 0xce, 0xc1, 0x22, 0x9d, 0x2e, 0x9f, 0xc6, 0x87, 0x08, 0xc1, 0x78, 0x14, 0x67, 0xa6,
	0xd6, 0x96, 0xae, 0xe5, 0xfd, 0x55, 0xeb, 0xcd, 0xdc, 0x8c, 0xee, 0x9d, 0xcb, 0x1f, 0xcd, 0xd2,
	0xc5, 0xcf, 0xa6, 0x43, 0x17, 0xce, 0x05, 0x6f, 0x51, 0xa3, 0x27, 0xf9, 0xcb, 0x31, 0xf4, 0xe7,
	0xcd, 0x60, 0x40, 0x41, 0x8e, 0x44, 0x2a, 0x01, 0xbf, 0x40, 0xd5, 0x0c, 0x64, 0x7e, 0xa2, 0xf4,
	0x25, 0x6b, 0x9d, 0x16, 0xb1, 0xf3, 0x28, 0xfc, 0x22, 0xda, 0x22, 0xeb, 0x17, 0xa1, 0x9a, 0x47,
	0x2d, 0x3f, 0xf8, 0x88, 0xdc, 0x35, 0x65, 0x57, 0xbc, 0x83, 0x65, 0xef, 0x00, 0xb7, 0xd1, 0x56,
	0x22, 0xb9, 0x74, 0xcb, 0xad, 0xca, 0x26, 0x63, 0xa8, 0x66, 0x04, 0x18, 0xdd, 0xbf, 0x1e, 0x89,
	0x51, 0x1b, 0x70, 0x9d, 0xa3, 0x70, 0x26, 0x8e, 0xe1, 0x7f, 0xe6, 0xd4, 0x44, 0xb5, 0x04, 0xd4,
	0x50, 0x0c, 0x8e, 0x52, 0x96, 0x80, 0x9e, 0xd2, 0x36, 0x45, 0x26, 0xf5, 0x9a, 0x25, 0x10, 0x3c,
	0x40, 0xbb, 0x0b, 0x8d, 0x4c, 0xf7, 0xce, 0xe7, 0x32, 0xaa, 0xf4, 0x24, 0xc7, 0xef, 0xd0, 0x2d,
	0x2d, 0x0b, 0x3f, 0x26, 0xeb, 0xf6, 0x9b, 0xac, 0x6c, 0x92, 0xf7, 0xe4, 0x26, 0x9a, 0x9d, 0xc5,
	0x29, 0xba, 0xbb, 0x6c, 0x27, 0x26, 0x1b, 0x4f, 0xae, 0xf5, 0xdd, 0x0b, 0xff, 0x99, 0x6f, 0x5b,
	0xbe, 0x47, 0x55, 0x73, 0x49, 0xbc, 0x59, 0xe4, 0x92, 0xdd, 0xde, 0xd3, 0x1b, 0x79, 0xa6, 0x74,
	0xf7, 0xf0, 0x72, 0xea, 0x3b, 0x57, 0x53, 0xdf, 0xf9, 0x35, 0xf5, 0x9d, 0x8b, 0x99, 0x5f, 0xba,
	0x9a, 0xf9, 0xa5, 0x6f, 0x33, 0xbf, 0xf4, 0xe1, 0x19, 0x8f, 0xd5, 0x30, 0x8f, 0x48, 0x5f, 0x24,
	0xf6, 0xeb, 0xb7, 0x8f, 0x3d, 0x39, 0x38, 0x0e, 0xc7, 0xf6, 0x67, 0xa2, 0x26, 0x23, 0x90, 0x51,
	0x55, 0x6f, 0xc6, 0xf3, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xe8, 0x5f, 0x9e, 0x5c, 0x69, 0x04,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// Grant grants the provided authorization to the grantee on the granter's
	// account with the provided expiration time.
	Grant(ctx context.Context, in *MsgGrantRequest, opts ...grpc.CallOption) (*MsgGrantResponse, error)
	// ExecAuthorized attempts to execute the provided messages using
	// authorizations granted to the grantee. Each message should have only
	// one signer corresponding to the granter of the authorization.
	ExecAuthorized(ctx context.Context, in *MsgExecRequest, opts ...grpc.CallOption) (*MsgExecResponse, error)
	// Revoke revokes any authorization corresponding to the provided method name on the
	// granter's account that has been granted to the grantee.
	Revoke(ctx context.Context, in *MsgRevokeRequest, opts ...grpc.CallOption) (*MsgRevokeResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) Grant(ctx context.Context, in *MsgGrantRequest, opts ...grpc.CallOption) (*MsgGrantResponse, error) {
	out := new(MsgGrantResponse)
	err := c.cc.Invoke(ctx, "/cosmos.authz.v1beta1.Msg/Grant", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ExecAuthorized(ctx context.Context, in *MsgExecRequest, opts ...grpc.CallOption) (*MsgExecResponse, error) {
	out := new(MsgExecResponse)
	err := c.cc.Invoke(ctx, "/cosmos.authz.v1beta1.Msg/ExecAuthorized", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Revoke(ctx context.Context, in *MsgRevokeRequest, opts ...grpc.CallOption) (*MsgRevokeResponse, error) {
	out := new(MsgRevokeResponse)
	err := c.cc.Invoke(ctx, "/cosmos.authz.v1beta1.Msg/Revoke", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// Grant grants the provided authorization to the grantee on the granter's
	// account with the provided expiration time.
	Grant(context.Context, *MsgGrantRequest) (*MsgGrantResponse, error)
	// Exec attempts to execute the provided messages using
	// authorizations granted to the grantee. Each message should have only
	// one signer corresponding to the granter of the authorization.
	Exec(context.Context, *MsgExecRequest) (*MsgExecResponse, error)
	// Revoke revokes any authorization corresponding to the provided method name on the
	// granter's account that has been granted to the grantee.
	Revoke(context.Context, *MsgRevokeRequest) (*MsgRevokeResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) Grant(ctx context.Context, req *MsgGrantRequest) (*MsgGrantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Grant not implemented")
}
func (*UnimplementedMsgServer) ExecAuthorized(ctx context.Context, req *MsgExecRequest) (*MsgExecResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecAuthorized not implemented")
}
func (*UnimplementedMsgServer) Revoke(ctx context.Context, req *MsgRevokeRequest) (*MsgRevokeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Revoke not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_Grant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgGrantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Grant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.authz.v1beta1.Msg/Grant",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Grant(ctx, req.(*MsgGrantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ExecAuthorized_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgExecRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Exec(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.authz.v1beta1.Msg/ExecAuthorized",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Exec(ctx, req.(*MsgExecRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Revoke_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Revoke(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.authz.v1beta1.Msg/Revoke",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Revoke(ctx, req.(*MsgRevokeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cosmos.authz.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Grant",
			Handler:    _Msg_Grant_Handler,
		},
		{
			MethodName: "ExecAuthorized",
			Handler:    _Msg_ExecAuthorized_Handler,
		},
		{
			MethodName: "Revoke",
			Handler:    _Msg_Revoke_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cosmos/authz/v1beta1/tx.proto",
}

func (m *MsgGrantRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgGrantRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgGrantRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.Expiration, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.Expiration):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintTx(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x22
	if m.Authorization != nil {
		{
			size, err := m.Authorization.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Grantee) > 0 {
		i -= len(m.Grantee)
		copy(dAtA[i:], m.Grantee)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Grantee)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Granter) > 0 {
		i -= len(m.Granter)
		copy(dAtA[i:], m.Granter)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Granter)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgExecResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgExecResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgExecResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Result != nil {
		{
			size, err := m.Result.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgExecRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgExecRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgExecRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Msgs) > 0 {
		for iNdEx := len(m.Msgs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Msgs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Grantee) > 0 {
		i -= len(m.Grantee)
		copy(dAtA[i:], m.Grantee)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Grantee)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgGrantResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgGrantResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgGrantResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgRevokeRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRevokeRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRevokeRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MsgTypeUrl) > 0 {
		i -= len(m.MsgTypeUrl)
		copy(dAtA[i:], m.MsgTypeUrl)
		i = encodeVarintTx(dAtA, i, uint64(len(m.MsgTypeUrl)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Grantee) > 0 {
		i -= len(m.Grantee)
		copy(dAtA[i:], m.Grantee)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Grantee)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Granter) > 0 {
		i -= len(m.Granter)
		copy(dAtA[i:], m.Granter)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Granter)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgRevokeResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRevokeResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRevokeResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgGrantRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Granter)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Grantee)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Authorization != nil {
		l = m.Authorization.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.Expiration)
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgExecResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Result != nil {
		l = m.Result.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgExecRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Grantee)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.Msgs) > 0 {
		for _, e := range m.Msgs {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *MsgGrantResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgRevokeRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Granter)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Grantee)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.MsgTypeUrl)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgRevokeResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgGrantRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgGrantRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgGrantRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Granter", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Granter = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Grantee", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Grantee = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authorization", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Authorization == nil {
				m.Authorization = &types.Any{}
			}
			if err := m.Authorization.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Expiration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.Expiration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgExecResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgExecAuthorizedResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgExecAuthorizedResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Result", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Result == nil {
				m.Result = &types1.Result{}
			}
			if err := m.Result.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgExecRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgExecAuthorizedRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgExecAuthorizedRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Grantee", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Grantee = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Msgs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Msgs = append(m.Msgs, &types.Any{})
			if err := m.Msgs[len(m.Msgs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgGrantResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgGrantResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgGrantResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgRevokeRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgRevokeRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRevokeRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Granter", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Granter = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Grantee", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Grantee = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MethodName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MsgTypeUrl = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgRevokeResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgRevokeResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRevokeResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
