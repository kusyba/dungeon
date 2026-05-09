package main

import (
	"flag"
	"log"
	"os"

	"dungeon/internal/config"
	"dungeon/internal/processor"
	"dungeon/internal/reader"
	"dungeon/internal/reporter"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	openTime, err := cfg.GetOpenTime()
	if err != nil {
		log.Fatalf("Failed to parse OpenAt: %v", err)
	}
	closeTime := cfg.GetCloseTime(openTime)

	events := reader.ReadEvents(os.Stdin, "events.txt")

	players, output := processor.ProcessEvents(events, cfg, openTime, closeTime)

	reporter.PrintOutput(output)
	reporter.PrintFinalReport(players, cfg)
}
