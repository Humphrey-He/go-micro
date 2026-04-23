package recommendation

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"go-micro/internal/recommendation/job"
)

type Scheduler struct {
	cron *cron.Cron
	db   *sqlx.DB
}

func NewScheduler(db *sqlx.DB) *Scheduler {
	return &Scheduler{
		cron: cron.New(),
		db:   db,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	// Association Job - every day at 01:00
	_, err := s.cron.AddFunc("0 1 * * *", func() {
		log.Println("[Scheduler] Running Association Job")
		j := job.NewAssociationJob(s.db)
		if err := j.Run(); err != nil {
			log.Printf("[Scheduler] Association Job failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule Association Job: %v", err)
	}

	// Item-CF Job - every day at 02:00
	_, err = s.cron.AddFunc("0 2 * * *", func() {
		log.Println("[Scheduler] Running Item-CF Job")
		j := job.NewItemCFJob(s.db)
		if err := j.Run(); err != nil {
			log.Printf("[Scheduler] Item-CF Job failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule Item-CF Job: %v", err)
	}

	// Bestseller Job - every day at 03:00
	_, err = s.cron.AddFunc("0 3 * * *", func() {
		log.Println("[Scheduler] Running Bestseller Job")
		j := job.NewBestsellerJob(s.db)
		if err := j.Run(); err != nil {
			log.Printf("[Scheduler] Bestseller Job failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule Bestseller Job: %v", err)
	}

	// Preference Job - every day at 04:00
	_, err = s.cron.AddFunc("0 4 * * *", func() {
		log.Println("[Scheduler] Running Preference Job")
		j := job.NewPreferenceJob(s.db)
		if err := j.Run(); err != nil {
			log.Printf("[Scheduler] Preference Job failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule Preference Job: %v", err)
	}

	log.Println("[Scheduler] Starting cron scheduler")
	s.cron.Start()

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("[Scheduler] Stopping cron scheduler")
	s.cron.Stop()
}
