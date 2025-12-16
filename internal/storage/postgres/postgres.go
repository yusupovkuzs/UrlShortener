package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go-url-shortener/internal/config"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string, dbInfo config.PostgressConfig) (*Storage, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", dbInfo.Host, dbInfo.Port, dbInfo.User, dbInfo.Password, dbInfo.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
        return nil, fmt.Errorf("%s: %w", psqlInfo, err)
    }
    defer db.Close()
	err = db.Ping()
    if err != nil {
        return nil, fmt.Errorf("%s: %w", psqlInfo, err)
    }
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT UNIQUE NOT NULL,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", psqlInfo, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", psqlInfo, err)
	}

	return &Storage{db: db}, nil
}
