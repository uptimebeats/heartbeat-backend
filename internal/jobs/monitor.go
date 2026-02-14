package jobs

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
	"github.com/uptimebeats/heartbeat-api/internal/repository"
)

type MonitorJob struct {
	Repo *repository.HeartbeatRepository
	Cron *cron.Cron
}

func NewMonitorJob(repo *repository.HeartbeatRepository) *MonitorJob {
	return &MonitorJob{
		Repo: repo,
		Cron: cron.New(),
	}
}

func (j *MonitorJob) Start() {
	// Run every 1 minute
	_, err := j.Cron.AddFunc("@every 1m", func() {
		log.Println("Running heartbeat failure check...")
		if err := j.Repo.CheckHeartbeatFailures(context.Background()); err != nil {
			log.Printf("Error checking heartbeat failures: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to schedule monitor job: %v", err)
	}

	j.Cron.Start()
	log.Println("Background monitor job started")
}

func (j *MonitorJob) Stop() {
	j.Cron.Stop()
}
