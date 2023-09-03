package gapi

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/myGo/simplebank/db/sqlc"
	pb "github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"github.com/myGo/simplebank/validate"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	if violations := validateLoginUserRequest(req); violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	user, err := server.store.GetUsers(ctx, req.Username)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "no user found in database: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "can not login the user: %s", err)
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "password mismatch: %s", err)
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "can not generate access token: %s", err)
	}

	refeshToken, refeshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokentDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can not generate refresh token: %s", err)
	}
	mtdt := server.extractMetadata(ctx)

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refeshPayload.ID,
		Username:     refeshPayload.Username,
		RefreshToken: refeshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiredAt:    pgtype.Timestamp{Time: refeshPayload.ExpriedAt},
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "can not create a session: %s", err)
	}

	rsp := &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refeshToken,
		AccessTokenExpriseAt:  timestamppb.New(accessPayload.ExpriedAt),
		RefreshTokenExpriseAt: timestamppb.New(refeshPayload.ExpriedAt),
	}
	return rsp, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validate.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, FieldViolation("Username", err))
	}

	if err := validate.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, FieldViolation("Password", err))
	}
	return violations
}
