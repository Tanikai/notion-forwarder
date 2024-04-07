package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log"
	"net/http"
	"notion-forwarder/dependencies"
	"notion-forwarder/handlers"
	"notion-forwarder/models"
)

var k = koanf.New(".")

func main() {
	// Load config
	if err := k.Load(file.Provider("./config.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
		return
	}
	var config models.NotionForwarderConfig
	if err := k.Unmarshal("", &config); err != nil {
		log.Fatalf("error loading config, could not unmarshal: %v", err)
		return
	}

	client := dependencies.NewNotionForwardingClient(config.IntegrationToken, config.Databases)

	if err := client.PopulateForwardedDatabases(); err != nil {
		log.Fatalf("error populating database: %v", err)
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello"))
		if err != nil {
			return
		}
	})
	r.Mount("/r", handlers.NotionForwarderRoutes(client))

	fmt.Printf("Starting server on :3000\n")
	if err := http.ListenAndServe(":3000", r); err != nil {
		return
	}
}
