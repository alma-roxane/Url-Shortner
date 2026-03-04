package main

import (
	"log"
	"net/http"

	"go-project/api-gateway/config"
	"go-project/api-gateway/routes"
)

func main() {
	cfg := config.Load()
	h := routes.New(cfg)

	log.Printf("api-gateway listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, h); err != nil {
		log.Fatal(err)
	}
}
