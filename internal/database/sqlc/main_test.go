// Package database provides database connection and query functionalities.
package database

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Sahas001/pay-on/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testQueries *Queries
	testPool    *pgxpool.Pool
)

func TestMain(m *testing.M) {
	var err error
	cfg, err := config.LoadConfig("../../..")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}
	ctx := context.Background()
	testPool, err = pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}
	defer testPool.Close()
	testQueries = New(testPool)
	os.Exit(m.Run())
}
