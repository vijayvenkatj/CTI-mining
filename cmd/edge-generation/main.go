package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/config"
	edgegen "github.com/vijayvenkatj/cti-miner/pkg/edge-gen"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rc := cfg.Kafka.EdgeGenerator.Reader
	if rc == nil {
		log.Fatal("kafka.edge_generator.reader is not configured")
	}
	reader, err := commons.NewKafkaReader(ctx, cfg.Kafka.Brokers, rc.Topic, rc.GroupID)
	if err != nil {
		log.Fatalf("kafka reader: %v", err)
	}
	defer reader.Close()

	wc := cfg.Kafka.EdgeGenerator.Writer
	if wc == nil {
		log.Fatal("kafka.edge_generator.writer is not configured")
	}
	writer, err := commons.NewKafkaWriter(ctx, cfg.Kafka.Brokers, wc.Topic)
	if err != nil {
		log.Fatalf("kafka writer: %v", err)
	}
	defer writer.Close()

	// 2^24 bits (2 MB): ~0.2% false positives at ~1M edges; 2^16 dropped >90% of edges (eval/results/bloom.csv).
	bloom := algorithms.NewBloomFilter(1<<24, 4)
	cms := algorithms.NewCMS(4, 1<<10)
	index := &resources.IndicatorIndex{}

	eg := edgegen.NewEdgeGenerator(bloom, cms, index, 1000, reader, writer)

	if err := eg.GenerateEdges(ctx); err != nil {
		log.Fatalf("generate edges: %v", err)
	}
}
