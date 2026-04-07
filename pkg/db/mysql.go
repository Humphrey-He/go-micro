package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go-micro/pkg/config"
)

func NewMySQL() (*sqlx.DB, error) {
	dsn := config.GetEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3307)/go_micro?parseTime=true&charset=utf8mb4&loc=Local")
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func Tx(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
