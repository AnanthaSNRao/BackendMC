package gapi

import (
	"context"

	"github.com/lib/pq"
	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	hashedPassword, err := util.HashedPassword(req.Password)

	if err != nil {
		ctx.Err()
		return nil, status.Errorf(codes.Internal, "failed to create a hash password: %s", err)
	}
	arg := db.CreateUserParams{
		Email:          req.GetEmail(),
		Username:       req.GetUsername(),
		FullName:       req.GetFullName(),
		HashedPassword: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":

				return nil, status.Errorf(codes.AlreadyExists, "User already exisits: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}
	resposnse := &pb.CreateUserResponse{
		User: convertUser(user),
	}
	return resposnse, nil

}
