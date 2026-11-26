package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	const fn = "storage.postgres.New"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) migrate() error {
	const fn = "storage.postgres.migrate"

	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL
	);
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const fn = "storage.postgres.SaveURL"

	res, err := s.db.Exec(`INSERT INTO url (url, alias) VALUES ($1, $2)`, urlToSave, alias)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, storage.ErrURLExists
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.postgres.GetURL"

	var url string
	err := s.db.QueryRow(`SELECT url FROM url WHERE alias = $1`, alias).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) (int64, error) {
	const fn = "storage.postgres.DeleteURL"

	res, err := s.db.Exec(`DELETE FROM url WHERE alias = $1`, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if rows == 0 {
		return 0, storage.ErrURLNotFound
	}

	return rows, nil
}

func (s *Storage) UpdateURL(alias, urlToSave string) error {
	const fn = "storage.postgres.UpdateURL"

	res, err := s.db.Exec(`UPDATE url SET url = $1 WHERE alias = $2`, urlToSave, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if rows == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}

func (s *Storage) AliasExists(alias string) (bool, error) {
	const fn = "storage.postgres.AliasExists"

	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM url WHERE alias = $1)`, alias).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	return exists, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}