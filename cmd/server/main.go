package main

import (
	"log"
	"net/http"

	"voidfs/internal/config"
	"voidfs/internal/server"
)

func main() {
	cfg := config.Load()
	app := server.New(cfg)

	log.Printf("starting server on %s with root %s", cfg.Addr, cfg.RootDir)
	if err := http.ListenAndServe(cfg.Addr, app.Router()); err != nil {
		log.Fatal(err)
	}
}
