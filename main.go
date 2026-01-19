package main

import (
	"context"
	"log"

	"github.com/Sahas001/pay-on/api"
	"github.com/Sahas001/pay-on/config"
	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	ctx := context.Background()
	conn, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}

	store := database.NewStore(conn)
	server := api.NewServer(cfg, store)
	err = server.Start(cfg.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start server:", err)
	}
}
