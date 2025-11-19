package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"final/bench"
	"final/model"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type PGConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func DefaultPGConfig() PGConfig {
	return PGConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "app",
		Password: "app",
		DBName:   "appdb",
		SSLMode:  "disable",
	}
}

func (c PGConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func NewPostgresDB(cfg PGConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed in sql.Open(): %w", err)
	}

	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return db, nil
}

func PingPostgres(db *sql.DB) {
	start := time.Now()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping failed: %v", err)
	}
	elapsed := time.Since(start)

	log.Printf("ping took: %v", elapsed)
}

func EnsureSchema(db *sql.DB) error {
	const query = `
	CREATE TABLE IF NOT EXISTS sessions(
	id TEXT PRIMARY KEY,
	session JSONB NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("query execution: %v", err)
	}
	return nil
}

func NewPgResult(name string, count int, total time.Duration) (bench.Result) {
	var avg time.Duration
	if count > 0 {
		avg = total / time.Duration(count)
	}
	return bench.Result{
		Name: name,
		Count: count,
		Total: total,
		AvgPerOp: avg,
	}
}

func InsertSingle(ctx context.Context, db *sql.DB, records []model.SessionRecord) (bench.Result, error) {
	name := "PG_INSERT_SINGLE"
	start := time.Now()

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO sessions (id, session)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			session = EXCLUDED.session`)
	if err != nil {
		return bench.Result{}, fmt.Errorf("failed to prepare context: %w", err)
	}

	for _, rec := range records {
		data, err := json.Marshal(rec)
		if err != nil {
			return bench.Result{}, fmt.Errorf("failed to marshal: %w", err)
		}

		if _, err := stmt.ExecContext(ctx, rec.ID, data); err != nil {
			return bench.Result{}, fmt.Errorf("failed to execute: %w", err)
		}

	}

	total := time.Since(start)
	var avg time.Duration
	if len(records) > 0 {
		avg = total / time.Duration(len(records))
	}
	result := bench.Result{
		Name:     name,
		Total:    total,
		Count:    len(records),
		AvgPerOp: avg,
	}
	return result, nil
}

func GetByID(ctx context.Context, db *sql.DB, ids []string) (bench.Result, error) {
	name := "PG_GET_BY_ID"
	start := time.Now()

	stmt, err := db.PrepareContext(ctx,
		"SELECT session FROM sessions WHERE id = $1")

	if err != nil {
		return bench.Result{}, fmt.Errorf("PrepareContext: %w", err)
	}
	defer stmt.Close()

	var rec model.SessionRecord

	for _, id := range ids {
		var raw []byte
		if err := stmt.QueryRowContext(ctx, id).Scan(&raw); err != nil {
			return bench.Result{}, fmt.Errorf("QueryRow.Scan: %w", err)
		}
		// We don't actually use 'rec', but this is how you'd unmarshal.
		if err := json.Unmarshal(raw, &rec); err != nil {
			return bench.Result{}, fmt.Errorf("json.Unmarshal: %w", err)
		}
	}

	total := time.Since(start)
	var avg time.Duration
	if len(ids) > 0 {
		avg = total / time.Duration(len(ids))
	}
	result := bench.Result{
		Name:     name,
		Count:    len(ids),
		Total:    total,
		AvgPerOp: avg,
	}
	return result, nil
}

func ConfigurePool(db *sql.DB, maxOpen int, maxIdle int) {
	if maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
	if maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	}
	db.SetConnMaxLifetime(0)
}

func InsertSingleConcurrent(ctx context.Context, db *sql.DB, records []model.SessionRecord, workers int) (bench.Result, error) {
	const name = "PG_INSERT_SINGLE_CONCURRENT"

	if workers <= 0 {
		workers = 1
	}
	totalCount := len(records)
	if totalCount == 0 {
		return NewPgResult(name, 0, 0), nil
	}

	chunkSize := (totalCount + workers - 1) / workers

	var wg sync.WaitGroup

	start := time.Now()

	for w := 0; w < workers; w++ {
		from := w * chunkSize
		if from >= totalCount {
			break
		}
		to := from + chunkSize
		if to > totalCount {
			to = totalCount
		}

		subset := records[from:to]

		wg.Add(1)
		go func (recs []model.SessionRecord)  {
			defer wg.Done()

			for _, rec := range recs {
				jsonData, err := json.Marshal(rec)
				if err != nil {
					return
				}

				_, err = db.ExecContext(ctx, "INSERT INTO sessions (id, session) VALUES ($1, $2)", 
				rec.ID, jsonData)
				if err != nil {
				}
			}
			
		} (subset)
	}
	wg.Wait()

	total := time.Since(start)
	return NewPgResult(name, totalCount, total), nil
}

func ClearSessions(db *sql.DB) error {
    _, err := db.Exec("TRUNCATE TABLE sessions")
    return err
}
