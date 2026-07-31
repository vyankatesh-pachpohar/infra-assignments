package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vyankatesh-pachpohar/infra-assignments/internal/domain"
)

var ErrNotFound = errors.New("config not found")

type ConfigRepository struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Ping verifies DB connectivity — used by the /ping health handler.
func (r *ConfigRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *ConfigRepository) GetByID(ctx context.Context, id string) (*domain.Config, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, host, port, app_name, log_level, updated_at FROM configs WHERE id = $1`, id)

	var c domain.Config
	if err := row.Scan(&c.ID, &c.Host, &c.Port, &c.AppName, &c.LogLevel, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// Upsert inserts a new config or updates it if the id already exists.
func (r *ConfigRepository) Upsert(ctx context.Context, c *domain.Config) error {
	c.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO configs (id, host, port, app_name, log_level, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			host = EXCLUDED.host,
			port = EXCLUDED.port,
			app_name = EXCLUDED.app_name,
			log_level = EXCLUDED.log_level,
			updated_at = EXCLUDED.updated_at
	`, c.ID, c.Host, c.Port, c.AppName, c.LogLevel, c.UpdatedAt)
	return err
}
