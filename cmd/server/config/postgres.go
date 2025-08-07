package config

import (
	"database/sql"
	"fmt"
)

func SetupDB(cfg *Config) (*sql.DB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.POSTGRES_USER,
		cfg.POSTGRES_PASSWORD,
		cfg.POSTGRES_HOSTNAME,
		cfg.POSTGRES_PORT,
		cfg.POSTGRES_DB,
		cfg.POSTGRES_SSL,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
