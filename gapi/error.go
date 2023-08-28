package gapi

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FieldViolation(err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       "Username",
		Description: err.Error(),
	}
}

func InvalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	badRequest := &errdetails.BadRequest{FieldViolations: violations}
	statusInavlid := status.New(codes.InvalidArgument, "invalid parameters")

	statusDetails, err := statusInavlid.WithDetails(badRequest)
	if err != nil {
		return statusInavlid.Err()
	}
	return statusDetails.Err()
}
