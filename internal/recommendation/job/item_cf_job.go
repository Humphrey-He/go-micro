package job

import (
	"log"
	"time"

	"ecomm-micro/internal/recommendation/algorithm"
	"github.com/jmoiron/sqlx"
)

type ItemCFJob struct {
	itemCF *algorithm.ItemCF
}

func NewItemCFJob(db *sqlx.DB) *ItemCFJob {
	return &ItemCFJob{
		itemCF: algorithm.NewItemCF(db),
	}
}

func (j *ItemCFJob) Run() error {
	log.Println("[ItemCFJob] starting...")
	start := time.Now()

	if err := j.itemCF.ComputeSimilarity(); err != nil {
		log.Printf("[ItemCFJob] failed: %v", err)
		return err
	}

	log.Printf("[ItemCFJob] completed in %v", time.Since(start))
	return nil
}
