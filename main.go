package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myGo/simplebank/api"
	db "github.com/myGo/simplebank/db/sqlc"
	_ "github.com/myGo/simplebank/doc/statik"
	"github.com/myGo/simplebank/gapi"
	"github.com/myGo/simplebank/pb"
	"github.com/myGo/simplebank/util"
	"github.com/myGo/simplebank/worker"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}
	// dbDriver := config.DBDriver
	dbSource := config.DBSource

	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	runDBMigrations(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistibutor(redisOpt)
	go runTaskProcessor(redisOpt, store)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)

}

func runDBMigrations(migrationUrl string, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up")
	}
}

func runGinServer(config util.Config, store db.Store) {
	serverAddress := config.HTTPServerAddress
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create a sever: %")
	}

	err = server.Start(serverAddress)

	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}

func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistibutor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server ")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcserver := grpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(grpcserver, server)
	reflection.Register(grpcserver)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}
	log.Printf("start gPRC server at %s", listener.Addr().String())

	err = grpcserver.Serve(listener)

	if err != nil {
		log.Fatal().Msg("cannot start grpc server")
	}
}

func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistibutor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server ")
	}

	grpcMux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal().Msg("cannot register handler server:")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()

	if err != nil {
		log.Fatal().Msg("cannot create handler server:")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create a lisnte")
	}

	err = http.Serve(listener, mux)
	log.Printf("start gateway server at %s", listener.Addr().String())

	if err != nil {
		log.Fatal().Msg("cannot start a HTTP gateway serve")
	}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store)

	log.Info().Msg("start task Processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Msg("Failed to start task Processor")
	}
}
