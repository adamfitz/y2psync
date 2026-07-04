package database

import (
	"database/sql"
)

type ConfigRepo struct {
	db *DB
}

func NewConfigRepo(db *DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

func (r *ConfigRepo) Set(key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

func (r *ConfigRepo) Get(key string) (string, error) {
	var value string
	err := r.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (r *ConfigRepo) Delete(key string) error {
	_, err := r.db.Exec(`DELETE FROM config WHERE key = ?`, key)
	return err
}
