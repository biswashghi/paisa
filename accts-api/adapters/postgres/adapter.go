package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"accts-api/adapters/postgres/internal/repository"
	"accts-api/ports"

	_ "github.com/lib/pq"
)

type Adapter struct {
	db *sql.DB
}

func New(ctx context.Context) (*Adapter, error) {
	db, err := openDB(ctx)
	if err != nil {
		return nil, err
	}
	schemaSQL, err := loadSchemaSQL()
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return nil, err
	}
	return &Adapter{db: db}, nil
}

func openDB(ctx context.Context) (*sql.DB, error) {
	psqlInfo, err := postgresDSNFromEnv()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func postgresDSNFromEnv() (string, error) {
	if dsn := os.Getenv("PAISA_DATABASE_URL"); dsn != "" {
		return dsn, nil
	}

	host := getenv("PAISA_POSTGRES_HOST", "localhost")
	portValue := getenv("PAISA_POSTGRES_PORT", "5243")
	port, err := strconv.Atoi(portValue)
	if err != nil {
		return "", fmt.Errorf("invalid PAISA_POSTGRES_PORT %q", portValue)
	}
	user := getenv("PAISA_POSTGRES_USER", "project")
	password := os.Getenv("PAISA_POSTGRES_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("PAISA_POSTGRES_PASSWORD is required when PAISA_DATABASE_URL is not set")
	}
	dbname := getenv("PAISA_POSTGRES_DB", "project")
	sslmode := getenv("PAISA_POSTGRES_SSLMODE", "disable")

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   dbname,
	}
	values := dsn.Query()
	values.Set("sslmode", sslmode)
	dsn.RawQuery = values.Encode()
	return dsn.String(), nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (a *Adapter) Close() error {
	return a.db.Close()
}

func (a *Adapter) Stores() ports.StoreSet {
	return repository.NewStoreSet(a.db)
}

func (a *Adapter) WithinTx(ctx context.Context, fn func(context.Context, ports.StoreSet) error) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.AppErrorFromDB(err)
	}
	defer tx.Rollback()

	if err := fn(ctx, repository.NewStoreSet(tx)); err != nil {
		return err
	}
	return repository.AppErrorFromDB(tx.Commit())
}
