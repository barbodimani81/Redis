package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type BenchResult struct {
	Name     string
	Count    int
	Total    time.Duration
	AvgPerOp time.Duration
}

type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Street  string `json:"street"`
	ZipCode string `json:"zip_code"`
}

type Event struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Metadata  string `json:"metadata"` // JSON-encoded string for extra data
}

type SessionRecord struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	UserName    string   `json:"user_name"`
	Email       string   `json:"email"`
	Device      string   `json:"device"`
	IP          string   `json:"ip"`
	IsMobile    bool     `json:"is_mobile"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   int64    `json:"created_at"`
	LastSeenAt  int64    `json:"last_seen_at"`
	Tags        []string `json:"tags"`
	Address     Address  `json:"address"`
	Events      []Event  `json:"events"`
	SessionNote string   `json:"session_note"`
}

func benchSetJSON(ctx context.Context, client *redis.Client, records []SessionRecord) (BenchResult, error) {
	start := time.Now()

	for _, rec := range records {
		data, err := json.Marshal(rec)
		if err != nil {
			return BenchResult{}, fmt.Errorf("marshal failed for id=%s: %w", rec.ID, err)
		}
		// Key pattern: session:set:<id>
		key := fmt.Sprintf("session:set:%s", rec.ID)

		if err := client.Set(ctx, key, data, 0).Err(); err != nil {
			return BenchResult{}, fmt.Errorf("redis SET failed for key=%s: %w", key, err)
		}
	}
	elapsed := time.Since(start)
	count := len(records)
	var avg time.Duration
	if count > 0 {
		avg = elapsed / time.Duration(count)
	}

	return BenchResult{
		Name:     "SET_JSON",
		Count:    count,
		Total:    elapsed,
		AvgPerOp: avg,
	}, nil

}

func generateSessionRecords(n int) []SessionRecord {
	records := make([]SessionRecord, n)
	now := time.Now().Unix()
	r := rand.New(rand.NewSource(42)) // deterministic

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("session-%d", i)

		addr := Address{
			City:    fmt.Sprintf("City-%d", i%50),
			Country: "Wonderland",
			Street:  fmt.Sprintf("Street %d", i),
			ZipCode: fmt.Sprintf("%05d", i%10000),
		}

		events := make([]Event, 0, 10)
		for j := 0; j < 10; j++ {
			ev := Event{
				Type:      fmt.Sprintf("event_type_%d", j%3),
				Timestamp: now - int64(j*60),
				Metadata: fmt.Sprintf(`{"action":"click","index":%d,"score":%d}`,
					j, r.Intn(1000)),
			}
			events = append(events, ev)
		}

		record := SessionRecord{
			ID:         id,
			UserID:     fmt.Sprintf("user-%d", i%200),
			UserName:   fmt.Sprintf("User Name %d", i),
			Email:      fmt.Sprintf("user%d@example.com", i),
			Device:     []string{"android", "ios", "web", "desktop"}[i%4],
			IP:         fmt.Sprintf("10.0.%d.%d", (i/256)%256, i%256),
			IsMobile:   i%2 == 0,
			IsActive:   i%3 != 0,
			CreatedAt:  now - int64(r.Intn(3600*24)),
			LastSeenAt: now - int64(r.Intn(3600)),
			Tags:       []string{"beta", "test", fmt.Sprintf("cohort-%d", i%5)},
			Address:    addr,
			Events:     events,
			SessionNote: fmt.Sprintf(
				"Some longer note about this session #%d with random value %d",
				i, r.Intn(1_000_000),
			),
		}

		records[i] = record
	}
	return records
}

func main() {
	ctx := context.Background()

	// For now we hardcode the address. Later we can move this to env/flags.
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // default DB
	})

	// Simple health check: PING
	start := time.Now()
	pong, err := client.Ping(ctx).Result()
	elapsed := time.Since(start)

	if err != nil {
		log.Fatalf("redis PING failed: %v", err)
	}

	fmt.Printf("PING response: %s (took %v)\n", pong, elapsed)

	sampleRecords := generateSessionRecords(3)
	sampleJSON, err := json.Marshal(sampleRecords[0])
	if err != nil {
		log.Fatalf("json marshal failed: %v", err)
	}

	fmt.Printf("sample jason size: %d bytes\n", len(sampleJSON))

	// Optionally print the JSON string once for inspection (can comment out later)
	fmt.Println("Sample JSON:", string(sampleJSON))

	// Prepare bigger batches for later phases (1000 and 10000)
	records1000 := generateSessionRecords(1000)
	records10000 := generateSessionRecords(10000)

	fmt.Printf("Generated %d records and %d records for later benchmarks.\n",
		len(records1000), len(records10000))

	if err := client.FlushDB(ctx).Err(); err != nil {
		log.Fatalf("FlushDB before 1000-run failed: %v", err)
	}

	fmt.Println("Running SET_JSON benchmark for 1000 records...")

	res1000, err := benchSetJSON(ctx, client, records1000)
	if err != nil {
		log.Fatalf("SET_JSON 1000 failed: %v", err)
	}

	fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
		res1000.Name, res1000.Count, res1000.Total, res1000.AvgPerOp)

	if err := client.FlushDB(ctx).Err(); err != nil {
		log.Fatalf("FlushDB before 10000-run failed: %v", err)
	}

	fmt.Println("Running SET_JSON benchmark for 10000 records...")

	res10000, err := benchSetJSON(ctx, client, records10000)
	if err != nil {
		log.Fatalf("SET_JSON 1000 failed: %v", err)
	}

	fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
		res10000.Name, res10000.Count, res10000.Total, res1000.AvgPerOp)

}
