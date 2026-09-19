# CTI-miner

A research implementation exploring streaming algorithms for large-scale
Cyber Threat Intelligence (CTI) processing. Indicators and pulses are
consumed from Kafka and analyzed using probabilistic data structures for
memory-efficient membership testing and frequency estimation.

## Structure

```
cmd/
  ingestion/        Polls OTX and publishes pulses to Kafka
  edge-generation/  Consumes pulses, publishes co-occurrence edges
pkg/
  algorithms/   Bloom filter, Count-Min Sketch
  commons/      Kafka reader/writer, HTTP client
  config/       Viper-based configuration loader
  edge-gen/     Edge generation pipeline
  otx/          AlienVault OTX client and poller
  resources/    CTI domain types (Pulse, Indicator, Edge)
```

## Configuration

Copy the example config and adjust as needed:

```bash
cp config.json.example config.json
```

```json
{
  "kafka": {
    "brokers": ["localhost:9092"],
    "reader": {
      "topic": "cti-pulses",
      "group_id": "cti-miner"
    },
    "writer": {
      "topic": "cti-edges"
    }
  },
  "otx": {
    "api_key": "your-otx-api-key"
  }
}
```

## Usage

```bash
go run ./cmd/ingestion
go run ./cmd/edge-generation
```

## Status

Early-stage research project. Algorithms and pipeline are under active
development.
