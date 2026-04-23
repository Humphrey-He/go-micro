package job

import (
	"log"
	"time"

	"go-micro/internal/recommendation/algorithm"
	"go-micro/pkg/db"
)

type BestsellerJob struct {
	bs *algorithm.Bestseller
}

func NewBestsellerJob() (*BestsellerJob, error) {
	dbx, err := db.NewMySQL()
	if err != nil {
		return nil, err
	}
	return &BestsellerJob{
		bs: algorithm.NewBestseller(dbx),
	}, nil
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
