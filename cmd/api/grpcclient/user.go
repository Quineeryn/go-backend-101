package grpcclient

import (
	"context"
	"time"

	userv1 "github.com/Quineeryn/go-backend-101/pkg/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type UserClient struct {
	cc *grpc.ClientConn
	US userv1.UserServiceClient
}

func NewUserClient(addr string) (*UserClient, error) {
	cc, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(3*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return &UserClient{cc: cc, US: userv1.NewUserServiceClient(cc)}, nil
}

func (c *UserClient) Close() error { return c.cc.Close() }

func WithRequestID(ctx context.Context, reqID string) context.Context {
	if reqID == "" {
		return ctx
	}
	md := metadata.New(map[string]string{"x-request-id": reqID})
	return metadata.NewOutgoingContext(ctx, md)
}
