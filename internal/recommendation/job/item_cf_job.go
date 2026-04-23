package job

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"go-micro/internal/recommendation/algorithm"
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