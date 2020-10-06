package service

import (
	"context"
	"grpc_youtube_tutorial/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer is server for authentication
type AuthServer struct {
	userStore  UserStore
	jwtManager *JWTManager
}

// NewAuthServer manager returns a pointer to a AuthServer
func NewAuthServer(userStore UserStore, jwtManager *JWTManager) *AuthServer {
	return &AuthServer{
		userStore:  userStore,
		jwtManager: jwtManager,
	}
}

// Login is a unary RPC to login user
func (server *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := server.userStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate access token: %v", err)
	}

	res := &pb.LoginResponse{AccessToken: token}
	return res, nil
}
