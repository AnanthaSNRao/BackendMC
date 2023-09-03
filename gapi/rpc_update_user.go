package gapi

import (
	"context"

	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"github.com/myGo/simplebank/validate"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := server.authoizeUser(ctx)

	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "Cannot update other user's info")
	}

	if violations := validateUpdateUserResponse(req); violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := util.HashedPassword(*req.Password)

		if err != nil {
			ctx.Err()
			return nil, status.Errorf(codes.Internal, "failed to create a hash password: %s", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}

		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}

	}

	user, err := server.store.UpdateUser(ctx, arg)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}
	resposnse := &pb.UpdateUserResponse{
		User: convertUser(user),
	}
	return resposnse, nil

}

func validateUpdateUserResponse(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validate.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, FieldViolation("username", err))
	}

	if req.FullName != nil {
		if err := validate.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, FieldViolation("Fullname", err))
		}

	}

	if req.Email != nil {
		if err := validate.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, FieldViolation("Email", err))
		}
	}

	if req.Password != nil {
		if err := validate.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, FieldViolation("Password", err))
		}
	}

	return violations
}
