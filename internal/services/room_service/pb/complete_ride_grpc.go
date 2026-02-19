package roomservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RoomServiceCompleteRideServer — расширение сервера с методом CompleteRide
// Реализуется вместе с RoomServiceServer
type RoomServiceCompleteRideServer interface {
	CompleteRide(context.Context, *CompleteRideRequest) (*CompleteRideResponse, error)
}

// CompleteRide добавляется к клиенту через extension method pattern
func CompleteRideClient(cc grpc.ClientConnInterface, ctx context.Context, req *CompleteRideRequest, opts ...grpc.CallOption) (*CompleteRideResponse, error) {
	out := new(CompleteRideResponse)
	err := cc.Invoke(ctx, "/service.room.v1.RoomService/CompleteRide", req, out,
		append([]grpc.CallOption{grpc.StaticMethod()}, opts...)...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CompleteRideHandler — handler для регистрации в grpc.Server
func CompleteRideHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompleteRideRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	impl, ok := srv.(RoomServiceCompleteRideServer)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "CompleteRide not implemented")
	}
	if interceptor == nil {
		return impl.CompleteRide(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/service.room.v1.RoomService/CompleteRide"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return impl.CompleteRide(ctx, req.(*CompleteRideRequest))
	})
}

// RegisterCompleteRide добавляет метод CompleteRide к уже запущенному gRPC серверу
func RegisterCompleteRide(s *grpc.Server, srv RoomServiceCompleteRideServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "service.room.v1.RoomService",
		HandlerType: (*RoomServiceCompleteRideServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CompleteRide",
				Handler:    CompleteRideHandler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "room.proto",
	}, srv)
}
