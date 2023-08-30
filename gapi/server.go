package gapi

import (
	"fmt"

	db "github.com/myGo/simplebank/db/sqlc"
	pb "github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/token"
	"github.com/myGo/simplebank/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

// creating a new gRPC server and setup routings
func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
