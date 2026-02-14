package main

import (
	"log"

	"github.com/uptimebeats/heartbeat-api/internal/api/router"
	"github.com/uptimebeats/heartbeat-api/internal/config"
	"github.com/uptimebeats/heartbeat-api/internal/database"
	"github.com/uptimebeats/heartbeat-api/internal/jobs"
	"github.com/uptimebeats/heartbeat-api/internal/repository"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Connect to Database
	dbPool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// 3. Initialize Repository
	repo := repository.NewHeartbeatRepository(dbPool)

	// 4. Start Background Jobs
	monitorJob := jobs.NewMonitorJob(repo)
	monitorJob.Start()
	defer monitorJob.Stop()

	// 5. Setup Router
	r := router.SetupRouter(repo)

	// 6. Start Server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil { // Listen and serve on 0.0.0.0:8080 (for windows/localhost check default)
		log.Fatalf("Failed to start server: %v", err)
	}
}
