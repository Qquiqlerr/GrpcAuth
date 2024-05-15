package auth

import (
	"context"
	ssov1 "github.com/Qquiqlerr/protos/gen/go"
	"github.com/asaskevich/govalidator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}
func (s *ServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponce, error) {
	//validating
	if !govalidator.IsEmail(req.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}
	if !govalidator.IsAlphanumeric(req.GetPassword()) || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	//register user and get userID
	ID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.RegisterResponce{
		UserId: ID,
	}, nil
}

func (s *ServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponce, error) {
	//validating
	if !govalidator.IsEmail(req.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}
	if !govalidator.IsAlphanumeric(req.GetPassword()) || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	if req.GetAppID() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "invalid app id")
	}

	//use service layer to get token
	jwt, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppID()))
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &ssov1.LoginResponce{
		Token: jwt,
	}, nil
}

func (s *ServerAPI) IsAdmin(cxt context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponce, error) {
	//validating
	if req.GetUserId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	//check is admin
	isAdmin, err := s.auth.IsAdmin(cxt, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &ssov1.IsAdminResponce{IsAdmin: isAdmin}, nil
}
