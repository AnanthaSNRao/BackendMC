package gapi

import (
	"fmt"

	db "github.com/myGo/simplebank/db/sqlc"
	pb "github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/token"
	"github.com/myGo/simplebank/util"
	"github.com/myGo/simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistibutor
}

// creating a new gRPC server and setup routings
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistibutor) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
