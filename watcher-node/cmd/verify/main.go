package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	path := flag.String("log", "logs/watcher_A.jsonl", "Path to jsonl log file")
	flag.Parse()
	total, valid, invalid, err := VerifyFile(*path)
	if err != nil { log.Fatalf("verify failed: %v", err) }
	fmt.Printf("verified %d entries: %d valid, %d invalid\n", total, valid, invalid)
	if invalid > 0 { fmt.Printf("some signatures invalid\n") }
}
