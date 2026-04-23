package job

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"go-micro/internal/recommendation/algorithm"
)

type BestsellerJob struct {
	bs *algorithm.Bestseller
}

func NewBestsellerJob(db *sqlx.DB) *BestsellerJob {
	return &BestsellerJob{
		bs: algorithm.NewBestseller(db),
	}
}

func (j *BestsellerJob) Run() error {
	log.Println("[BestsellerJob] starting...")
	start := time.Now()

	// Compute category bestsellers
	if err := j.bs.ComputeCategoryBestsellers(); err != nil {
		log.Printf("[BestsellerJob] category bestsellers failed: %v", err)
	}

	// Compute global bestsellers
	if err := j.bs.ComputeGlobalBestsellers(); err != nil {
		log.Printf("[BestsellerJob] global bestsellers failed: %v", err)
	}

	log.Printf("[BestsellerJob] completed in %v", time.Since(start))
	return nil
}
