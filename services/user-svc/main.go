package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	userv1 "github.com/Quineeryn/go-backend-101/pkg/proto/user/v1"
	"github.com/Quineeryn/go-backend-101/services/user-svc/server"
	"github.com/Quineeryn/go-backend-101/services/user-svc/store"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	st := store.NewMemStore()
	userv1.RegisterUserServiceServer(s, server.NewUserServer(st))

	hs := health.NewServer()
	healthpb.RegisterHealthServer(s, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus(userv1.UserService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	log.Printf("user-svc gRPC on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
