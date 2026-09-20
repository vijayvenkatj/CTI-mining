package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/config"
	"github.com/vijayvenkatj/cti-miner/pkg/estimator"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rc := cfg.Kafka.Estimator.Reader
	if rc == nil {
		log.Fatal("kafka.estimator.reader is not configured")
	}
	reader, err := commons.NewKafkaReader(ctx, cfg.Kafka.Brokers, rc.Topic, rc.GroupID)
	if err != nil {
		log.Fatalf("kafka reader: %v", err)
	}
	defer reader.Close()

	triest := algorithms.NewTreist(10)

	est := estimator.NewEstimator(triest, reader)

	if err := est.Run(ctx); err != nil {
		log.Fatalf("run estimator: %v", err)
	}
}
