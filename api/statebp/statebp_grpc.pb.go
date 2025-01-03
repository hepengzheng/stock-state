// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.29.2
// source: statebp.proto

package statepb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	State_GetStock_FullMethodName = "/statepb.State/GetStock"
)

// StateClient is the client API for State service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StateClient interface {
	GetStock(ctx context.Context, opts ...grpc.CallOption) (State_GetStockClient, error)
}

type stateClient struct {
	cc grpc.ClientConnInterface
}

func NewStateClient(cc grpc.ClientConnInterface) StateClient {
	return &stateClient{cc}
}

func (c *stateClient) GetStock(ctx context.Context, opts ...grpc.CallOption) (State_GetStockClient, error) {
	stream, err := c.cc.NewStream(ctx, &State_ServiceDesc.Streams[0], State_GetStock_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &stateGetStockClient{stream}
	return x, nil
}

type State_GetStockClient interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type stateGetStockClient struct {
	grpc.ClientStream
}

func (x *stateGetStockClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *stateGetStockClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// StateServer is the server API for State service.
// All implementations must embed UnimplementedStateServer
// for forward compatibility
type StateServer interface {
	GetStock(State_GetStockServer) error
	mustEmbedUnimplementedStateServer()
}

// UnimplementedStateServer must be embedded to have forward compatible implementations.
type UnimplementedStateServer struct {
}

func (UnimplementedStateServer) GetStock(State_GetStockServer) error {
	return status.Errorf(codes.Unimplemented, "method GetStock not implemented")
}
func (UnimplementedStateServer) mustEmbedUnimplementedStateServer() {}

// UnsafeStateServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StateServer will
// result in compilation errors.
type UnsafeStateServer interface {
	mustEmbedUnimplementedStateServer()
}

func RegisterStateServer(s grpc.ServiceRegistrar, srv StateServer) {
	s.RegisterService(&State_ServiceDesc, srv)
}

func _State_GetStock_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StateServer).GetStock(&stateGetStockServer{stream})
}

type State_GetStockServer interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type stateGetStockServer struct {
	grpc.ServerStream
}

func (x *stateGetStockServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *stateGetStockServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// State_ServiceDesc is the grpc.ServiceDesc for State service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var State_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "statepb.State",
	HandlerType: (*StateServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetStock",
			Handler:       _State_GetStock_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "statebp.proto",
}