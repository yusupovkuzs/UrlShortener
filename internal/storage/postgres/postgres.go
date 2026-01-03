package postgres

import (
	// project
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/storage"

	// embedded
	"database/sql"
	"fmt"

	// external 
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbInfo config.PostgresConfig) (*Storage, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbInfo.Host, dbInfo.Port, dbInfo.User, dbInfo.Password, dbInfo.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	stmt := "CREATE TABLE IF NOT EXISTS url(id BIGSERIAL PRIMARY KEY, alias TEXT UNIQUE NOT NULL, url TEXT NOT NULL); CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);"

	_, err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	_, err := s.db.Exec("INSERT INTO url(url, alias) VALUES ($1, $2)", urlToSave, alias)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			return fmt.Errorf("%w", storage.ErrURlExists)
		}
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	var url string
	err := s.db.QueryRow("SELECT url FROM url WHERE alias=$1", alias).Scan(&url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%w", storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%w", err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
    res, err := s.db.Exec("DELETE FROM url WHERE alias=$1", alias)
    if err != nil {
        return fmt.Errorf("%w", err)
    }
    ra, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("%w", err)
    }
    if ra == 0 {
        return fmt.Errorf("%w", storage.ErrURLNotFound)
    }
    return nil
}
