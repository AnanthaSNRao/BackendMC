package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"github.com/myGo/simplebank/validate"
	"github.com/myGo/simplebank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if violations := validateCreateUserRequest(req); violations != nil {
		return nil, InvalidArgumentError(violations)
	}
	hashedPassword, err := util.HashedPassword(req.Password)

	if err != nil {
		ctx.Err()
		return nil, status.Errorf(codes.Internal, "failed to create a hash password: %s", err)
	}
	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Email:          req.GetEmail(),
			Username:       req.GetUsername(),
			FullName:       req.GetFullName(),
			HashedPassword: hashedPassword,
		},

		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)

		},
	}

	txRes, err := server.store.CreateUserTx(ctx, arg)

	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {

			return nil, status.Errorf(codes.AlreadyExists, "User already exisits: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}
	// todo: db transaction

	resposnse := &pb.CreateUserResponse{
		User: convertUser(txRes.User),
	}
	return resposnse, nil

}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validate.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, FieldViolation("username", err))
	}

	if err := validate.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, FieldViolation("Fullname", err))
	}

	if err := validate.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, FieldViolation("Email", err))
	}

	if err := validate.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, FieldViolation("Password", err))
	}
	return violations
}
