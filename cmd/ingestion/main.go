package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/config"
	"github.com/vijayvenkatj/cti-miner/pkg/otx"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	modifiedSince, err := time.Parse(time.RFC3339, cfg.OTX.ModifiedSince)
	if err != nil {
		log.Fatalf("parse otx.modified_since: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Publishes pulses to the same topic edge-generation reads from.
	writer, err := commons.NewKafkaWriter(ctx, cfg.Kafka.Brokers, cfg.Kafka.Reader.Topic)
	if err != nil {
		log.Fatalf("kafka writer: %v", err)
	}
	defer writer.Close()

	client := otx.NewClient(cfg.OTX.APIKey, commons.NewHTTPClient(http.DefaultClient))
	poller := otx.NewPoller(cfg.OTX.BaseURL, modifiedSince, cfg.OTX.InitialBackoff, cfg.OTX.MaxBackoff, client, writer)

	poller.Run(ctx)
}
