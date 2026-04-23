package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB is a wrapper around sqlx.DB
type DB struct {
	*sqlx.DB
}

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

// New creates a new database connection
func New(cfg *Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode, cfg.Timezone,
	)

	// Debug: log connection string (mask password)
	log.Printf("DB config - Host: '%s', Port: '%s', User: '%s', DB: '%s', SSLMode: '%s', Timezone: '%s'",
		cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.SSLMode, cfg.Timezone)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database successfully")
	return &DB{db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	d.DB.Close()
	return nil
}

// BeginTx starts a transaction
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return d.DB.BeginTxx(ctx, opts)
}

// Exec executes a query without returning rows
func (d *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.DB.ExecContext(ctx, query, args...)
}

// Query executes a query and returns rows
func (d *DB) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return d.DB.QueryxContext(ctx, query, args...)
}

// QueryRow executes a query and returns a single row
func (d *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return d.DB.QueryRowxContext(ctx, query, args...)
}

// Get queries a single row into a struct
func (d *DB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return d.DB.GetContext(ctx, dest, query, args...)
}

// Select queries multiple rows into a slice
func (d *DB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return d.DB.SelectContext(ctx, dest, query, args...)
}

// NamedQuery executes a named query
func (d *DB) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return d.DB.NamedQueryContext(ctx, query, arg)
}

// NamedExec executes a named exec
func (d *DB) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return d.DB.NamedExecContext(ctx, query, arg)
}
