// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// CallingClient is the client API for Calling service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CallingClient interface {
	VoIP(ctx context.Context, opts ...grpc.CallOption) (Calling_VoIPClient, error)
}

type callingClient struct {
	cc grpc.ClientConnInterface
}

func NewCallingClient(cc grpc.ClientConnInterface) CallingClient {
	return &callingClient{cc}
}

func (c *callingClient) VoIP(ctx context.Context, opts ...grpc.CallOption) (Calling_VoIPClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Calling_serviceDesc.Streams[0], "/Calling/VoIP", opts...)
	if err != nil {
		return nil, err
	}
	x := &callingVoIPClient{stream}
	return x, nil
}

type Calling_VoIPClient interface {
	Send(*ClientREQ) error
	Recv() (*ServerRES, error)
	grpc.ClientStream
}

type callingVoIPClient struct {
	grpc.ClientStream
}

func (x *callingVoIPClient) Send(m *ClientREQ) error {
	return x.ClientStream.SendMsg(m)
}

func (x *callingVoIPClient) Recv() (*ServerRES, error) {
	m := new(ServerRES)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CallingServer is the server API for Calling service.
// All implementations must embed UnimplementedCallingServer
// for forward compatibility
type CallingServer interface {
	VoIP(Calling_VoIPServer) error
	mustEmbedUnimplementedCallingServer()
}

// UnimplementedCallingServer must be embedded to have forward compatible implementations.
type UnimplementedCallingServer struct {
}

func (UnimplementedCallingServer) VoIP(Calling_VoIPServer) error {
	return status.Errorf(codes.Unimplemented, "method VoIP not implemented")
}
func (UnimplementedCallingServer) mustEmbedUnimplementedCallingServer() {}

// UnsafeCallingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CallingServer will
// result in compilation errors.
type UnsafeCallingServer interface {
	mustEmbedUnimplementedCallingServer()
}

func RegisterCallingServer(s *grpc.Server, srv CallingServer) {
	s.RegisterService(&_Calling_serviceDesc, srv)
}

func _Calling_VoIP_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CallingServer).VoIP(&callingVoIPServer{stream})
}

type Calling_VoIPServer interface {
	Send(*ServerRES) error
	Recv() (*ClientREQ, error)
	grpc.ServerStream
}

type callingVoIPServer struct {
	grpc.ServerStream
}

func (x *callingVoIPServer) Send(m *ServerRES) error {
	return x.ServerStream.SendMsg(m)
}

func (x *callingVoIPServer) Recv() (*ClientREQ, error) {
	m := new(ClientREQ)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Calling_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Calling",
	HandlerType: (*CallingServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "VoIP",
			Handler:       _Calling_VoIP_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "plot.proto",
}
