package job

import (
	"log"

	"ecomm-micro/internal/recommendation/algorithm"
	"github.com/jmoiron/sqlx"
)

type AssociationJob struct {
	assoc *algorithm.Association
}

func NewAssociationJob(db *sqlx.DB) *AssociationJob {
	return &AssociationJob{
		assoc: algorithm.NewAssociation(db),
	}
}

// Run computes association rules from purchase data
func (j *AssociationJob) Run() error {
	log.Println("[AssociationJob] Computing association rules...")
	if err := j.assoc.ComputeAssociationRules(); err != nil {
		log.Printf("[AssociationJob] Failed to compute association rules: %v", err)
		return err
	}
	log.Println("[AssociationJob] Association rules computed successfully")
	return nil
}
