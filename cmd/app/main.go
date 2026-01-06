package main

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/virgiliusnanamanek02/job-scheduler/internal/job"
	"github.com/virgiliusnanamanek02/job-scheduler/internal/worker"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/virgiliusnanamanek02/job-scheduler/internal/config"
)

func main() {

	appConfig := config.NewConfig()

	log.Printf("Loading configuration from:")
	log.Printf("  - Env File: %s", appConfig.EnvFile)
	log.Printf("  - DB Config: %s", appConfig.DBConfigFile)
	log.Printf("  - Migration Dir: %s", appConfig.MigrationDir)

	db, err := config.ConnectDB(appConfig.DBConfigFile)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	err = config.ApplyMigrationWithConfig(db, appConfig)
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("Application started successfully")

	switch os.Getenv("APP_MODE") {
	case "worker":
		repo := job.NewRepository(db)
		w := worker.New(
			repo,
			30*time.Second, // lease duration
			2*time.Second,  // poll interval
		)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		w.Run(ctx)

	default:
		log.Fatalf("unsupported APP_MODE")
	}
}
