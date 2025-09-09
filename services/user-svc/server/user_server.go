package server

import (
	"context"
	"time"

	userv1 "github.com/Quineeryn/go-backend-101/pkg/proto/user/v1"
	"github.com/Quineeryn/go-backend-101/services/user-svc/store"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	st *store.MemStore
}

func NewUserServer(st *store.MemStore) *UserServer { return &UserServer{st: st} }

func (s *UserServer) GetUser(ctx context.Context, in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if u, ok := s.st.Get(in.GetId()); ok {
		return &userv1.GetUserResponse{User: toProto(u)}, nil
	}
	return nil, status.Error(codes.NotFound, "user not found")
}

func (s *UserServer) CreateUser(ctx context.Context, in *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	id := uuid.NewString()
	u := store.User{ID: id, Email: in.GetEmail(), Name: in.GetName(), CreatedAt: time.Now()}
	u = s.st.Create(u)
	return &userv1.CreateUserResponse{User: toProto(u)}, nil
}

func (s *UserServer) ListUsers(ctx context.Context, in *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users := s.st.List()
	out := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		out = append(out, toProto(u))
	}
	return &userv1.ListUsersResponse{Users: out}, nil
}

func toProto(u store.User) *userv1.User {
	return &userv1.User{
		Id: u.ID, Email: u.Email, Name: u.Name,
		CreatedAtUnix: u.CreatedAt.Unix(),
	}
}
