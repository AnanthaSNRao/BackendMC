package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/token"
	"github.com/myGo/simplebank/util"
	"github.com/myGo/simplebank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistibutor) *Server {
	config := util.Config{
		TokenSymetricKey:    util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationTypeBearer, accessToken)
	md := metadata.MD{
		authorizationHeaderKey: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashedPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}
