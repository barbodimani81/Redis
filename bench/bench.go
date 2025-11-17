package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"final/model"
)

type Result struct {
	Name     string
	Count    int
	Total    time.Duration
	AvgPerOp time.Duration
}

func benchHeader(name string, count int, total time.Duration) Result {
	var avg time.Duration
	if count > 0 {
		avg = total / time.Duration(count)
	}
	return Result{
		Name:     name,
		Count:    count,
		Total:    total,
		AvgPerOp: avg,
	}
}

func SetJSON(ctx context.Context, client *redis.Client, records []model.SessionRecord) (Result, error) {
	start := time.Now()

	for _, rec := range records {
		data, err := json.Marshal(rec)
		if err != nil {
			return Result{}, fmt.Errorf("marshal failed for id=%s: %w", rec.ID, err)
		}
		key := fmt.Sprintf("session:set:%s", rec.ID)

		if err := client.Set(ctx, key, data, 0).Err(); err != nil {
			return Result{}, fmt.Errorf("redis SET failed for key=%s: %w", key, err)
		}
	}

	elapsed := time.Since(start)
	return benchHeader("SET_JSON", len(records), elapsed), nil
}

func GetJSON(ctx context.Context, client *redis.Client, records []model.SessionRecord) (Result, error) {
	start := time.Now()

	for _, rec := range records {
		key := fmt.Sprintf("session:set:%s", rec.ID)

		data, err := client.Get(ctx, key).Bytes()
		if err != nil {
			return Result{}, fmt.Errorf("redis GET failed for key=%s: %w", key, err)
		}

		var out model.SessionRecord
		if err := json.Unmarshal(data, &out); err != nil {
			return Result{}, fmt.Errorf("json unmarshal failed for key=%s: %w", key, err)
		}
	}

	elapsed := time.Since(start)
	return benchHeader("GET_JSON", len(records), elapsed), nil
}

func HSet(ctx context.Context, client *redis.Client, records []model.SessionRecord) (Result, error) {
	start := time.Now()

	for _, rec := range records {
		hash, err := rec.ToHash()
		if err != nil {
			return Result{}, fmt.Errorf("ToHash failed for id=%s: %w", rec.ID, err)
		}

		key := fmt.Sprintf("session:hash:%s", rec.ID)

		if err := client.HSet(ctx, key, hash).Err(); err != nil {
			return Result{}, fmt.Errorf("redis HSET failed for key=%s: %w", key, err)
		}
	}

	elapsed := time.Since(start)
	return benchHeader("HSET_HASH", len(records), elapsed), nil
}

func HGetAll(ctx context.Context, client *redis.Client, records []model.SessionRecord) (Result, error) {
	start := time.Now()

	for _, rec := range records {
		key := fmt.Sprintf("session:hash:%s", rec.ID)

		hash, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			return Result{}, fmt.Errorf("redis HGETALL failed for key=%s: %w", key, err)
		}
		if len(hash) == 0 {
			return Result{}, fmt.Errorf("empty hash for key=%s", key)
		}

		if _, err := model.FromHash(hash); err != nil {
			return Result{}, fmt.Errorf("FromHash failed for key=%s: %w", key, err)
		}
	}

	elapsed := time.Since(start)
	return benchHeader("HGETALL_HASH", len(records), elapsed), nil
}

func SetJSONPipeline(ctx context.Context, client *redis.Client, records []model.SessionRecord) (Result, error) {
	start := time.Now()

	pipe := client.Pipeline()

	for _, rec := range records {
		data, err := json.Marshal(rec)
		if err != nil {
			return Result{}, fmt.Errorf("marshal failed for id=%s: %w", rec.ID, err)
		}
		key := fmt.Sprintf("session:set:%s", rec.ID)

		pipe.Set(ctx, key, data, 0)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("pipeline SET exec failed: %w", err)
	}

	elapsed := time.Since(start)
	return benchHeader("SET_JSON_PIPELINE", len(records), elapsed), nil
}

func HSetPipeline(ctx context.Context, client *redis.Client, records []model.SessionRecord) (Result, error) {
	start := time.Now()

	pipe := client.Pipeline()

	for _, rec := range records {
		hash, err := rec.ToHash()
		if err != nil {
			return Result{}, fmt.Errorf("ToHash failed for id=%s: %w", rec.ID, err)
		}

		key := fmt.Sprintf("session:hash:%s", rec.ID)

		pipe.HSet(ctx, key, hash)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("pipeline HSET exec failed: %w", err)
	}

	elapsed := time.Since(start)
	return benchHeader("HSET_HASH_PIPELINE", len(records), elapsed), nil
}
