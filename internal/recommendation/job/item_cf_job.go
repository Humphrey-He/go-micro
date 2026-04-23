package job

import (
	"log"
	"time"

	"go-micro/internal/recommendation/algorithm"
	"go-micro/pkg/db"
)

type ItemCFJob struct {
	itemCF *algorithm.ItemCF
}

func NewItemCFJob() (*ItemCFJob, error) {
	dbx, err := db.NewMySQL()
	if err != nil {
		return nil, err
	}
	return &ItemCFJob{
		itemCF: algorithm.NewItemCF(dbx),
	}, nil
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