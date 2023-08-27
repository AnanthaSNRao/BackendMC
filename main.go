package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq"
	"github.com/myGo/simplebank/api"
	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/gapi"
	"github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	dbDriver := config.DBDriver
	dbSource := config.DBSource

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	store := db.NewStore(conn)
	runGrpcServer(config, store)

}

func runGinServer(config util.Config, store db.Store) {
	serverAddress := config.HTTPServerAddress
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create a sever: %w", err)
	}

	err = server.Start(serverAddress)

	if err != nil {
		log.Fatal("cannot start server")
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server :", err)
	}

	grpcserver := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcserver, server)
	reflection.Register(grpcserver)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal("cannot create listener")
	}
	log.Printf("start gPRC server at %s", listener.Addr().String())

	err = grpcserver.Serve(listener)

	if err != nil {
		log.Fatal("cannot start grpc server")
	}
}
