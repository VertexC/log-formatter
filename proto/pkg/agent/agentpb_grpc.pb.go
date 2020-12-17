// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package agent

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// LogFormatterAgentClient is the client API for LogFormatterAgent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogFormatterAgentClient interface {
	ChangeConfig(ctx context.Context, in *ChangeConfigRequest, opts ...grpc.CallOption) (*ChangeConfigResponse, error)
	GetHeartBeat(ctx context.Context, in *HeartBeatRequest, opts ...grpc.CallOption) (*HeartBeat, error)
}

type logFormatterAgentClient struct {
	cc grpc.ClientConnInterface
}

func NewLogFormatterAgentClient(cc grpc.ClientConnInterface) LogFormatterAgentClient {
	return &logFormatterAgentClient{cc}
}

func (c *logFormatterAgentClient) ChangeConfig(ctx context.Context, in *ChangeConfigRequest, opts ...grpc.CallOption) (*ChangeConfigResponse, error) {
	out := new(ChangeConfigResponse)
	err := c.cc.Invoke(ctx, "/agentpb.LogFormatterAgent/ChangeConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logFormatterAgentClient) GetHeartBeat(ctx context.Context, in *HeartBeatRequest, opts ...grpc.CallOption) (*HeartBeat, error) {
	out := new(HeartBeat)
	err := c.cc.Invoke(ctx, "/agentpb.LogFormatterAgent/GetHeartBeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogFormatterAgentServer is the server API for LogFormatterAgent service.
// All implementations must embed UnimplementedLogFormatterAgentServer
// for forward compatibility
type LogFormatterAgentServer interface {
	ChangeConfig(context.Context, *ChangeConfigRequest) (*ChangeConfigResponse, error)
	GetHeartBeat(context.Context, *HeartBeatRequest) (*HeartBeat, error)
	mustEmbedUnimplementedLogFormatterAgentServer()
}

// UnimplementedLogFormatterAgentServer must be embedded to have forward compatible implementations.
type UnimplementedLogFormatterAgentServer struct {
}

func (UnimplementedLogFormatterAgentServer) ChangeConfig(context.Context, *ChangeConfigRequest) (*ChangeConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeConfig not implemented")
}
func (UnimplementedLogFormatterAgentServer) GetHeartBeat(context.Context, *HeartBeatRequest) (*HeartBeat, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHeartBeat not implemented")
}
func (UnimplementedLogFormatterAgentServer) mustEmbedUnimplementedLogFormatterAgentServer() {}

// UnsafeLogFormatterAgentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogFormatterAgentServer will
// result in compilation errors.
type UnsafeLogFormatterAgentServer interface {
	mustEmbedUnimplementedLogFormatterAgentServer()
}

func RegisterLogFormatterAgentServer(s grpc.ServiceRegistrar, srv LogFormatterAgentServer) {
	s.RegisterService(&_LogFormatterAgent_serviceDesc, srv)
}

func _LogFormatterAgent_ChangeConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogFormatterAgentServer).ChangeConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agentpb.LogFormatterAgent/ChangeConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogFormatterAgentServer).ChangeConfig(ctx, req.(*ChangeConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogFormatterAgent_GetHeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartBeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogFormatterAgentServer).GetHeartBeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agentpb.LogFormatterAgent/GetHeartBeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogFormatterAgentServer).GetHeartBeat(ctx, req.(*HeartBeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LogFormatterAgent_serviceDesc = grpc.ServiceDesc{
	ServiceName: "agentpb.LogFormatterAgent",
	HandlerType: (*LogFormatterAgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChangeConfig",
			Handler:    _LogFormatterAgent_ChangeConfig_Handler,
		},
		{
			MethodName: "GetHeartBeat",
			Handler:    _LogFormatterAgent_GetHeartBeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "agent/agentpb.proto",
}
