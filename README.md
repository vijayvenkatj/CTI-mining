# CTI-miner

A research implementation exploring streaming algorithms for large-scale
Cyber Threat Intelligence (CTI) processing. Indicators and pulses are
consumed from Kafka and analyzed using probabilistic data structures for
memory-efficient membership testing, frequency estimation, and triangle
counting.

## Structure

```
cmd/
  ingestion/        Polls OTX and publishes pulses to Kafka
  edge-generation/  Consumes pulses, publishes co-occurrence edges
  estimator/        Consumes edges, estimates triangle counts (TRIEST)
  server/           HTTP API over stored pulses/indicators/edges
pkg/
  algorithms/   Bloom filter, Count-Min Sketch, TRIEST
  commons/      Kafka reader/writer, HTTP client
  config/       Viper-based configuration loader
  edge-gen/     Edge generation pipeline
  estimator/    Triangle-estimation pipeline
  http/         HTTP router + controllers
  otx/          AlienVault OTX client and poller
  resources/    CTI domain types (Pulse, Indicator, Edge) + pagination
  storage/      Postgres persistence (optional)
migrations/     SQL schema migrations
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
    "edge_generator": {
      "reader": { "topic": "cti-pulses", "group_id": "cti-miner" },
      "writer": { "topic": "cti-edges" }
    },
    "estimator": {
      "reader": { "topic": "cti-edges", "group_id": "cti-miner-estimator" }
    }
  },
  "otx": {
    "api_key": "your-otx-api-key"
  },
  "postgres": {
    "dsn": "postgres://cti:cti@localhost:5433/ctiminer?sslmode=disable"
  },
  "server": {
    "addr": ":8080"
  }
}
```

`postgres` is optional — omit it and persistence/ingestion-state resume are
simply skipped. `server.addr` defaults to `:8080` if omitted.

## Usage

Local infra (Kafka, Kafka UI, Postgres):

```bash
docker compose up -d kafka kafka-ui postgres
```

Then run each service:

```bash
go run ./cmd/ingestion
go run ./cmd/edge-generation
go run ./cmd/estimator
go run ./cmd/server
```

## Docker

Runs the entire pipeline in containers, exposing only Kafka UI and the API:

```bash
cp config.docker.json.example config.docker.json   # fill in your OTX api_key
docker compose up -d --build
```

- Kafka UI: `http://localhost:8081`
- API: `http://localhost:8082`

Kafka, Postgres, and the ingestion/edge-generation/estimator pipeline stay
internal to the compose network.

## API

All endpoints are read-only and support `?page=&limit=` pagination.

| Endpoint | Filter | Notes |
|---|---|---|
| `GET /edges` | `pulse_id` | edges touching a pulse |
| `GET /pulses` | `include_indicators=true` | nests indicators per pulse |
| `GET /pulses/{id}` | — | always includes indicators |
| `GET /indicators` | `type` | e.g. `IPv4`, `domain` |
| `GET /indicators/{id}` | — | |

## Status

Early-stage research project. Algorithms and pipeline are under active
development.
