package grpcclient

import (
	"context"
	"log"
	"time"

	userv1 "github.com/Quineeryn/go-backend-101/pkg/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type UserClient struct {
	cc *grpc.ClientConn
	US userv1.UserServiceClient
}

func NewUserClient(addr string) (*UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cc, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingPolicy":"round_robin",
			"methodConfig":[{
				"name":[{"service":"user.v1.UserService"}],
				"retryPolicy":{
					"MaxAttempts":4,
					"InitialBackoff":"0.2s",
					"MaxBackoff":"2s",
					"BackoffMultiplier":2.0,
					"RetryableStatusCodes":["UNAVAILABLE","RESOURCE_EXHAUSTED"]
				}
			}]
		}`),
		grpc.WithUnaryInterceptor(unaryClientLogger),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &UserClient{cc: cc, US: userv1.NewUserServiceClient(cc)}, nil
}

func unaryClientLogger(
	ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("[gRPC] method=%s dur=%s err=%v", method, time.Since(start), err)
	return err
}

func (c *UserClient) Close() error { return c.cc.Close() }

func WithRequestID(ctx context.Context, reqID string) context.Context {
	if reqID == "" {
		return ctx
	}
	md := metadata.New(map[string]string{"x-request-id": reqID})
	return metadata.NewOutgoingContext(ctx, md)
}
