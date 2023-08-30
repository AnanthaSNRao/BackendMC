package gapi

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	startTime := time.Now()
	res, err := handler(ctx, req)
	endTime := time.Since(startTime)

	statusCode := codes.Unknown

	if st, ok := status.FromError(err); !ok {
		statusCode = st.Code()
	}
	logger := log.Info()

	if err != nil {
		logger = log.Error().Err(err)
	}
	logger.Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Dur("Duration", endTime).
		Str("Status_text", statusCode.String()).
		Int("Status_code", int(statusCode)).
		Msg("recived a grpc request")
	return res, err
}
