package repository

import (
	"database/sql"
	"fmt"
	"os"

	"misaki/config"

	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	logger *zap.Logger
	conn   *sql.DB
}

func NewSQLite(config *config.Config, logger *zap.Logger) (Repository, error) {
	configDb := config.Database
	db, err := sql.Open(
		"sqlite3",
		fmt.Sprintf("%s%s?_foreign_keys=on&cache=%s", configDb.Type, configDb.Address, configDb.Cache),
	)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open connections
	db.SetMaxOpenConns(configDb.MaxConn)

	// Ping to check if the database connection is established
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	repo := &SQLite{
		conn:   db,
		logger: logger,
	}

	err = repo.migrate(configDb.Schema)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *SQLite) migrate(filepath string) error {
	// Read the schema file
	schema, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Execute the SQL statements from the schema file
	_, err = s.conn.Exec(string(schema))
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite) Something() {
	log := s.logger.Sugar()
	log.Infow("Im on repository brruhhh")
}
