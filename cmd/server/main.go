package main

import (
	"log"
	"net/http"

	"github.com/vijayvenkatj/cti-miner/pkg/config"
	apihttp "github.com/vijayvenkatj/cti-miner/pkg/http"
	"github.com/vijayvenkatj/cti-miner/pkg/storage"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.Postgres == nil {
		log.Fatal("postgres is not configured")
	}

	pg, err := storage.NewPostgres(cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	router := apihttp.NewRouter(pg)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
