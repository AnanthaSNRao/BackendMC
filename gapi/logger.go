package gapi

import (
	"context"
	"net/http"
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

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()

		handler.ServeHTTP(res, req)

		logger.Str("protocol", "grpc").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Dur("Duration", duration).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Int("status_code", rec.StatusCode).
			Msg("recived a Http request")
	})
}
