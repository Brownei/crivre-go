package main

import (
	"log"

	"github.com/brownei/crivre-go/cmd/api"
	"github.com/brownei/crivre-go/db"
	"github.com/brownei/crivre-go/store"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error with the logger: %s", err.Error())
	}

	newDb, err := db.NewPostgresStorage()
	if err != nil {
		log.Fatalf("Error with the logger: %s", err.Error())
	}

	zapLogger := logger.Sugar()
	store := store.NewStore(newDb)
	db.InitializeDb(newDb)
	db.AddMigrations(newDb)

	server := api.NewApplication(zapLogger, store)

	if err := server.Run(); err != nil {
		log.Printf("Error Running: %s", err.Error())
	}
}
