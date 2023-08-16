package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/myGo/simplebank/api"
	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	dbDriver := config.DBDriver
	dbSource := config.DBSource
	serverAddress := config.ServerAddress

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	store := db.NewStore(conn)

	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create a sever: %w", err)
	}

	err = server.Start(serverAddress)

	if err != nil {
		log.Fatal("cannot start server")
	}
}
