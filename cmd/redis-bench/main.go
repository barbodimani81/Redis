package main

import (
	"context"
	"encoding/json"
	"final/bench"
	"final/model"
	"final/store"
	"fmt"
	"log"
)

func printResult(r bench.Result) {
	fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
		r.Name, r.Count, r.Total, r.AvgPerOp)
}

func main() {
	ctx := context.Background()

	client := store.NewClient()
	store.Ping(ctx, client)

	sample := model.GenerateSessionRecords(3)
	sampleJSON, err := json.Marshal(sample[0])
	if err != nil {
		log.Fatalf("sample json marshal failed: %v", err)
	}
	fmt.Printf("sample json size: %d bytes\n", len(sampleJSON))

	records1000 := model.GenerateSessionRecords(1000)
	records10000 := model.GenerateSessionRecords(10000)
	fmt.Printf("Generated %d and %d records\n", len(records1000), len(records10000))

	// SET / GET – 1000
	store.Flush(ctx, client)
	fmt.Println("Running SET_JSON 1000...")
	set1000, err := bench.SetJSON(ctx, client, records1000)
	if err != nil {
		log.Fatalf("SET_JSON 1000 failed: %v", err)
	}
	printResult(set1000)

	fmt.Println("Running GET_JSON 1000...")
	get1000, err := bench.GetJSON(ctx, client, records1000)
	if err != nil {
		log.Fatalf("GET_JSON 1000 failed: %v", err)
	}
	printResult(get1000)

	// SET / GET – 10000
	store.Flush(ctx, client)
	fmt.Println("Running SET_JSON 10000...")
	set10000, err := bench.SetJSON(ctx, client, records10000)
	if err != nil {
		log.Fatalf("SET_JSON 10000 failed: %v", err)
	}
	printResult(set10000)

	fmt.Println("Running GET_JSON 10000...")
	get10000, err := bench.GetJSON(ctx, client, records10000)
	if err != nil {
		log.Fatalf("GET_JSON 10000 failed: %v", err)
	}
	printResult(get10000)

	// HSET / HGETALL – 1000
	store.Flush(ctx, client)
	fmt.Println("Running HSET_HASH 1000...")
	hset1000, err := bench.HSet(ctx, client, records1000)
	if err != nil {
		log.Fatalf("HSET_HASH 1000 failed: %v", err)
	}
	printResult(hset1000)

	fmt.Println("Running HGETALL_HASH 1000...")
	hget1000, err := bench.HGetAll(ctx, client, records1000)
	if err != nil {
		log.Fatalf("HGETALL_HASH 1000 failed: %v", err)
	}
	printResult(hget1000)

	// HSET / HGETALL – 10000
	store.Flush(ctx, client)
	fmt.Println("Running HSET_HASH 10000...")
	hset10000, err := bench.HSet(ctx, client, records10000)
	if err != nil {
		log.Fatalf("HSET_HASH 10000 failed: %v", err)
	}
	printResult(hset10000)

	fmt.Println("Running HGETALL_HASH 10000...")
	hget10000, err := bench.HGetAll(ctx, client, records10000)
	if err != nil {
		log.Fatalf("HGETALL_HASH 10000 failed: %v", err)
	}
	printResult(hget10000)

	store.Flush(ctx, client)
	fmt.Println("Running SET_JSON_PIPELINE 10000...")
	setPipe10000, err := bench.SetJSONPipeline(ctx, client, records10000)
	if err != nil {
		log.Fatalf("SET_JSON_PIPELINE 10000 failed: %v", err)
	}
	printResult(setPipe10000)

	// HSET hash pipeline
	store.Flush(ctx, client)
	fmt.Println("Running HSET_HASH_PIPELINE 10000...")
	hsetPipe10000, err := bench.HSetPipeline(ctx, client, records10000)
	if err != nil {
		log.Fatalf("HSET_HASH_PIPELINE 10000 failed: %v", err)
	}
	printResult(hsetPipe10000)
}
