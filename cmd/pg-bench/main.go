package main

import (
	"context"
	"final/model"
	"final/store"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	count := flag.Int("count", 1000, "number of records to insert/select")
	mode := flag.String("mode", "insert", "mode: insert | get | both")
	maxOpenConns := flag.Int("max-open-conns", 4, "max open connections in Postgres pool")
	maxIdleConns := flag.Int("max-idle-conns", 4, "max idle connections in Postgres pool")
	workers := flag.Int("workers", 4, "number of concurrent workers (for concurrent modes)")
	batchSize := flag.Int("batch-size", 100, "batch size for batch insert mode")
	flag.Parse()

	ctx := context.Background()

	cfg := store.DefaultPGConfig()
	db, err := store.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("failed new db: %v", err)
	}

	store.ConfigurePool(db, *maxOpenConns, *maxIdleConns)

	defer db.Close()
	if err := store.EnsureSchema(db); err != nil {
		log.Fatalf("ensure schema failed: %v", err)
	}

	store.PingPostgres(db)

	if err := store.ClearSessions(db); err != nil {
		log.Fatalf("ClearSessions: %v", err)
	}
	
	records := model.GenerateSessionRecords(*count)

	switch *mode {
	case "insert":
		r, err := store.InsertSingle(ctx, db, records)
		if err != nil {
			log.Fatalf("InsertSingle: %v", err)
		}
		fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
			r.Name, r.Count, r.Total, r.AvgPerOp)

	case "get":
		// assume records already inserted; we just reuse their IDs
		ids := make([]string, len(records))
		for i, rec := range records {
			ids[i] = rec.ID
		}
		r, err := store.GetByID(ctx, db, ids)
		if err != nil {
			log.Fatalf("GetSessionsByID: %v", err)
		}
		fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
			r.Name, r.Count, r.Total, r.AvgPerOp)

	case "both":
		r1, err := store.InsertSingle(ctx, db, records)
		if err != nil {
			log.Fatalf("InsertSingle: %v", err)
		}
		fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
			r1.Name, r1.Count, r1.Total, r1.AvgPerOp)

		ids := make([]string, len(records))
		for i, rec := range records {
			ids[i] = rec.ID
		}
		// tiny sleep just to avoid mixing logs
		time.Sleep(500 * time.Millisecond)

		r2, err := store.GetByID(ctx, db, ids)
		if err != nil {
			log.Fatalf("GetByID: %v", err)
		}
		fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
			r2.Name, r2.Count, r2.Total, r2.AvgPerOp)

	case "insert_concurrent":
		r, err := store.InsertSingleConcurrent(ctx, db, records, *workers)
    if err != nil {
        log.Fatalf("InsertSingleConcurrent: %v", err)
    }
    fmt.Printf("[%s] workers=%d maxOpenConns=%d count=%d total=%v avg/op=%v\n",
        r.Name, *workers, *maxOpenConns, r.Count, r.Total, r.AvgPerOp)


	case "insert_batch":
		// Optional but recommended: clear table before each run
		if err := store.ClearSessions(db); err != nil {
			log.Fatalf("ClearSessions: %v", err)
		}
	
		r, err := store.InsertBatchMultiRow(ctx, db, records, *batchSize)
		if err != nil {
			log.Fatalf("InsertBatchMultiRow: %v", err)
		}
	
		fmt.Printf("[%s] batchSize=%d count=%d total=%v avg/op=%v\n",
			r.Name, *batchSize, r.Count, r.Total, r.AvgPerOp)

			
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}