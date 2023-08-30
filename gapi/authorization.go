package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/myGo/simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func (server *Server) authoizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	values := md.Get(authorizationHeaderKey)
	if len(values) == 0 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authHeader := values[0]

	fields := strings.Fields(authHeader)

	if len(fields) < 2 {
		return nil, fmt.Errorf("inavlid authrization format")
	}

	authType := strings.ToLower(fields[0])

	if authType != authorizationTypeBearer {
		return nil, fmt.Errorf("inavlid authrization Type")
	}

	accessToken := fields[1]

	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("inavlid access token: %s", err)
	}

	return payload, nil
}
