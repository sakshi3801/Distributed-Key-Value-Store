package kv

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// KeyValueClient is the client API for KeyValue service.
type KeyValueClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	Replicate(ctx context.Context, in *ReplicateRequest, opts ...grpc.CallOption) (*ReplicateResponse, error)
}

type keyValueClient struct {
	cc grpc.ClientConnInterface
}

func NewKeyValueClient(cc grpc.ClientConnInterface) KeyValueClient {
	return &keyValueClient{cc}
}

func (c *keyValueClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/kv.KeyValue/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValueClient) Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error) {
	out := new(SetResponse)
	err := c.cc.Invoke(ctx, "/kv.KeyValue/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValueClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/kv.KeyValue/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValueClient) Replicate(ctx context.Context, in *ReplicateRequest, opts ...grpc.CallOption) (*ReplicateResponse, error) {
	out := new(ReplicateResponse)
	err := c.cc.Invoke(ctx, "/kv.KeyValue/Replicate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeyValueServer is the server API for KeyValue service.
type KeyValueServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Set(context.Context, *SetRequest) (*SetResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	Replicate(context.Context, *ReplicateRequest) (*ReplicateResponse, error)
}

type UnimplementedKeyValueServer struct{}

func (UnimplementedKeyValueServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedKeyValueServer) Set(context.Context, *SetRequest) (*SetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedKeyValueServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedKeyValueServer) Replicate(context.Context, *ReplicateRequest) (*ReplicateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Replicate not implemented")
}

func RegisterKeyValueServer(s grpc.ServiceRegistrar, srv KeyValueServer) {
	s.RegisterService(&KeyValue_ServiceDesc, srv)
}

func _KeyValue_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.KeyValue/Get"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValue_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.KeyValue/Set"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueServer).Set(ctx, req.(*SetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValue_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.KeyValue/Delete"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValue_Replicate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplicateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueServer).Replicate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.KeyValue/Replicate"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueServer).Replicate(ctx, req.(*ReplicateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var KeyValue_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kv.KeyValue",
	HandlerType: (*KeyValueServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Get", Handler: _KeyValue_Get_Handler},
		{MethodName: "Set", Handler: _KeyValue_Set_Handler},
		{MethodName: "Delete", Handler: _KeyValue_Delete_Handler},
		{MethodName: "Replicate", Handler: _KeyValue_Replicate_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/kv.proto",
}

// ClusterClient is the client API for Cluster service.
type ClusterClient interface {
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	RequestVote(ctx context.Context, in *RequestVoteRequest, opts ...grpc.CallOption) (*RequestVoteResponse, error)
	GetClusterState(ctx context.Context, in *GetClusterStateRequest, opts ...grpc.CallOption) (*GetClusterStateResponse, error)
}

type clusterClient struct {
	cc grpc.ClientConnInterface
}

func NewClusterClient(cc grpc.ClientConnInterface) ClusterClient {
	return &clusterClient{cc}
}

func (c *clusterClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, "/kv.Cluster/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) RequestVote(ctx context.Context, in *RequestVoteRequest, opts ...grpc.CallOption) (*RequestVoteResponse, error) {
	out := new(RequestVoteResponse)
	err := c.cc.Invoke(ctx, "/kv.Cluster/RequestVote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterClient) GetClusterState(ctx context.Context, in *GetClusterStateRequest, opts ...grpc.CallOption) (*GetClusterStateResponse, error) {
	out := new(GetClusterStateResponse)
	err := c.cc.Invoke(ctx, "/kv.Cluster/GetClusterState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClusterServer is the server API for Cluster service.
type ClusterServer interface {
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	RequestVote(context.Context, *RequestVoteRequest) (*RequestVoteResponse, error)
	GetClusterState(context.Context, *GetClusterStateRequest) (*GetClusterStateResponse, error)
}

type UnimplementedClusterServer struct{}

func (UnimplementedClusterServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedClusterServer) RequestVote(context.Context, *RequestVoteRequest) (*RequestVoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedClusterServer) GetClusterState(context.Context, *GetClusterStateRequest) (*GetClusterStateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClusterState not implemented")
}

func RegisterClusterServer(s grpc.ServiceRegistrar, srv ClusterServer) {
	s.RegisterService(&Cluster_ServiceDesc, srv)
}

func _Cluster_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.Cluster/Heartbeat"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cluster_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestVoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.Cluster/RequestVote"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServer).RequestVote(ctx, req.(*RequestVoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cluster_GetClusterState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClusterStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServer).GetClusterState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kv.Cluster/GetClusterState"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServer).GetClusterState(ctx, req.(*GetClusterStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Cluster_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kv.Cluster",
	HandlerType: (*ClusterServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Heartbeat", Handler: _Cluster_Heartbeat_Handler},
		{MethodName: "RequestVote", Handler: _Cluster_RequestVote_Handler},
		{MethodName: "GetClusterState", Handler: _Cluster_GetClusterState_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/kv.proto",
}
