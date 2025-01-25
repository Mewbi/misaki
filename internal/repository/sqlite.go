package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"misaki/config"
	"misaki/types"

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

func (s *SQLite) CreateUser(ctx context.Context, user *types.User) error {
	query := `INSERT INTO users (id, telegram_id, telegram_name, admin, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.conn.Exec(query,
		user.UserID,
		user.TelegramID,
		user.TelegramName,
		user.Admin,
		user.CreatedAt,
	)
	return err
}

func (s *SQLite) GetUser(ctx context.Context, user *types.User) (*types.User, error) {
	query := `SELECT id, telegram_id, telegram_name, admin, created_at FROM users WHERE id = $1 OR telegram_id = $2`
	err := s.conn.QueryRow(query, user.UserID, user.TelegramID).Scan(
		&user.UserID,
		&user.TelegramID,
		&user.TelegramName,
		&user.Admin,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *SQLite) DeleteUser(ctx context.Context, user *types.User) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	query := `DELETE FROM users WHERE id = $1 OR telegram_id = $2`
	_, err = tx.Exec(query, user.UserID, user.TelegramID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite) GetBilling(ctx context.Context, billing *types.Billing) (*types.Billing, error) {
	return nil, nil
}

func (s *SQLite) ListBillings(ctx context.Context) ([]*types.Billing, error) {
	return nil, nil
}

func (s *SQLite) CreateBilling(ctx context.Context, billing *types.Billing) error {
	return nil
}

func (s *SQLite) DeleteBilling(ctx context.Context, billing *types.Billing) error {
	return nil
}

func (s *SQLite) AssociateBilling(ctx context.Context, payment *types.Payment) error {
	return nil
}

func (s *SQLite) ChangePaymentBilling(ctx context.Context, payment *types.Payment) error {
	return nil
}
