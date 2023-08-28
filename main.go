package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/myGo/simplebank/api"
	db "github.com/myGo/simplebank/db/sqlc"
	_ "github.com/myGo/simplebank/doc/statik"
	"github.com/myGo/simplebank/gapi"
	"github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"github.com/rakyll/statik/fs"
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
	go runGatewayServer(config, store)
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

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server :", err)
	}

	grpcMux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal("cannot register handler server: ", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()

	if err != nil {
		log.Fatal("cannot create handler server:")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create a lisnter", err)
	}

	err = http.Serve(listener, mux)
	log.Printf("start gateway server at %s", listener.Addr().String())

	if err != nil {
		log.Fatal("cannot start a HTTP gateway server", err)
	}

}
