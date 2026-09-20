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

	log.Println("listening on", cfg.Server.Addr)
	if err := http.ListenAndServe(cfg.Server.Addr, router); err != nil {
		log.Fatal(err)
	}
}
