// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// MetaFetcherClient is the client API for MetaFetcher service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetaFetcherClient interface {
	Get(ctx context.Context, in *MetaRequest, opts ...grpc.CallOption) (*Meta, error)
}

type metaFetcherClient struct {
	cc grpc.ClientConnInterface
}

func NewMetaFetcherClient(cc grpc.ClientConnInterface) MetaFetcherClient {
	return &metaFetcherClient{cc}
}

func (c *metaFetcherClient) Get(ctx context.Context, in *MetaRequest, opts ...grpc.CallOption) (*Meta, error) {
	out := new(Meta)
	err := c.cc.Invoke(ctx, "/imdb2meta.MetaFetcher/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetaFetcherServer is the server API for MetaFetcher service.
// All implementations must embed UnimplementedMetaFetcherServer
// for forward compatibility
type MetaFetcherServer interface {
	Get(context.Context, *MetaRequest) (*Meta, error)
	mustEmbedUnimplementedMetaFetcherServer()
}

// UnimplementedMetaFetcherServer must be embedded to have forward compatible implementations.
type UnimplementedMetaFetcherServer struct {
}

func (UnimplementedMetaFetcherServer) Get(context.Context, *MetaRequest) (*Meta, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedMetaFetcherServer) mustEmbedUnimplementedMetaFetcherServer() {}

// UnsafeMetaFetcherServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetaFetcherServer will
// result in compilation errors.
type UnsafeMetaFetcherServer interface {
	mustEmbedUnimplementedMetaFetcherServer()
}

func RegisterMetaFetcherServer(s grpc.ServiceRegistrar, srv MetaFetcherServer) {
	s.RegisterService(&MetaFetcher_ServiceDesc, srv)
}

func _MetaFetcher_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetaFetcherServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/imdb2meta.MetaFetcher/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetaFetcherServer).Get(ctx, req.(*MetaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MetaFetcher_ServiceDesc is the grpc.ServiceDesc for MetaFetcher service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetaFetcher_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "imdb2meta.MetaFetcher",
	HandlerType: (*MetaFetcherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _MetaFetcher_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}
