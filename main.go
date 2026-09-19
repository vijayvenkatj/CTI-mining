package main

import (
	"context"
	"log"

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

	reader := commons.NewKafkaReader(cfg.Kafka.Brokers, cfg.Kafka.Reader.Topic, cfg.Kafka.Reader.GroupID)
	defer reader.Close()

	writer := commons.NewKafkaWriter(cfg.Kafka.Brokers, cfg.Kafka.Writer.Topic)
	defer writer.Close()

	bloom := algorithms.NewBloomFilter(1<<16, 4)
	cms := algorithms.NewCMS(4, 1<<10)
	index := &resources.IndicatorIndex{}

	eg := edgegen.NewEdgeGenerator(bloom, cms, index, 1000, reader, writer)

	if err := eg.GenerateEdges(context.Background()); err != nil {
		log.Fatalf("generate edges: %v", err)
	}
}
