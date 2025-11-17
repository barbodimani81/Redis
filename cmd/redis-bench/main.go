package main

import (
	"context"
	"encoding/json"
	"final/bench"
	"final/model"
	"final/store"
	"flag"
	"fmt"
	"log"
)

func printResult(r bench.Result) {
	fmt.Printf("[%s] count=%d total=%v avg/op=%v\n",
		r.Name, r.Count, r.Total, r.AvgPerOp)
}

func main() {
	// ---- Flags ----
	count := flag.Int("count", 10000, "number of records to benchmark (e.g. 1000 or 10000)")
	pattern := flag.String("pattern", "both", "which pattern to run: set | hash | both")
	pipeline := flag.Bool("pipeline", false, "use pipeline for writes")
	withReads := flag.Bool("reads", true, "also benchmark reads (GET / HGETALL)")
	flag.Parse()

	ctx := context.Background()

	client := store.NewClient()
	store.Ping(ctx, client)

	// Generate data
	records := model.GenerateSessionRecords(*count)
	fmt.Printf("Generated %d records\n", len(records))

	// Show sample JSON size once
	sampleJSON, err := json.Marshal(records[0])
	if err != nil {
		log.Fatalf("sample json marshal failed: %v", err)
	}
	fmt.Printf("Sample JSON size: %d bytes\n", len(sampleJSON))

	runSet := *pattern == "set" || *pattern == "both"
	runHash := *pattern == "hash" || *pattern == "both"

	// ---- SET / GET path ----
	if runSet {
		store.Flush(ctx, client)

		if *pipeline {
			fmt.Println("Running SET_JSON_PIPELINE...")
			res, err := bench.SetJSONPipeline(ctx, client, records)
			if err != nil {
				log.Fatalf("SET_JSON_PIPELINE failed: %v", err)
			}
			printResult(res)
		} else {
			fmt.Println("Running SET_JSON...")
			res, err := bench.SetJSON(ctx, client, records)
			if err != nil {
				log.Fatalf("SET_JSON failed: %v", err)
			}
			printResult(res)
		}

		if *withReads {
			fmt.Println("Running GET_JSON...")
			res, err := bench.GetJSON(ctx, client, records)
			if err != nil {
				log.Fatalf("GET_JSON failed: %v", err)
			}
			printResult(res)
		}
	}

	// ---- HSET / HGETALL path ----
	if runHash {
		store.Flush(ctx, client)

		if *pipeline {
			fmt.Println("Running HSET_HASH_PIPELINE...")
			res, err := bench.HSetPipeline(ctx, client, records)
			if err != nil {
				log.Fatalf("HSET_HASH_PIPELINE failed: %v", err)
			}
			printResult(res)
		} else {
			fmt.Println("Running HSET_HASH...")
			res, err := bench.HSet(ctx, client, records)
			if err != nil {
				log.Fatalf("HSET_HASH failed: %v", err)
			}
			printResult(res)
		}

		if *withReads {
			fmt.Println("Running HGETALL_HASH...")
			res, err := bench.HGetAll(ctx, client, records)
			if err != nil {
				log.Fatalf("HGETALL_HASH failed: %v", err)
			}
			printResult(res)
		}
	}
}
