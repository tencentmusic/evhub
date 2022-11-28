// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.11.0
// source: transaction/transaction.proto

package transaction

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

// EvhubTransactionClient is the client API for EvhubTransaction service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EvhubTransactionClient interface {
	Callback(ctx context.Context, in *CallbackReq, opts ...grpc.CallOption) (*CallbackRsp, error)
}

type evhubTransactionClient struct {
	cc grpc.ClientConnInterface
}

func NewEvhubTransactionClient(cc grpc.ClientConnInterface) EvhubTransactionClient {
	return &evhubTransactionClient{cc}
}

func (c *evhubTransactionClient) Callback(ctx context.Context, in *CallbackReq, opts ...grpc.CallOption) (*CallbackRsp, error) {
	out := new(CallbackRsp)
	err := c.cc.Invoke(ctx, "/evhub_transaction.evhubTransaction/Callback", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EvhubTransactionServer is the server API for EvhubTransaction service.
// All implementations should embed UnimplementedEvhubTransactionServer
// for forward compatibility
type EvhubTransactionServer interface {
	Callback(context.Context, *CallbackReq) (*CallbackRsp, error)
}

// UnimplementedEvhubTransactionServer should be embedded to have forward compatible implementations.
type UnimplementedEvhubTransactionServer struct {
}

func (UnimplementedEvhubTransactionServer) Callback(context.Context, *CallbackReq) (*CallbackRsp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Callback not implemented")
}

// UnsafeEvhubTransactionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EvhubTransactionServer will
// result in compilation errors.
type UnsafeEvhubTransactionServer interface {
	mustEmbedUnimplementedEvhubTransactionServer()
}

func RegisterEvhubTransactionServer(s grpc.ServiceRegistrar, srv EvhubTransactionServer) {
	s.RegisterService(&EvhubTransaction_ServiceDesc, srv)
}

func _EvhubTransaction_Callback_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CallbackReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EvhubTransactionServer).Callback(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/evhub_transaction.evhubTransaction/Callback",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EvhubTransactionServer).Callback(ctx, req.(*CallbackReq))
	}
	return interceptor(ctx, in, info, handler)
}

// EvhubTransaction_ServiceDesc is the grpc.ServiceDesc for EvhubTransaction service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EvhubTransaction_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "evhub_transaction.evhubTransaction",
	HandlerType: (*EvhubTransactionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Callback",
			Handler:    _EvhubTransaction_Callback_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "transaction/transaction.proto",
}
