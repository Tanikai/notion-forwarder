package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log/slog"
	"net/http"
	"notion-forwarder/dependencies"
	"notion-forwarder/handlers"
	"notion-forwarder/models"
)

var k = koanf.New(".")

func main() {
	// Load config
	if err := k.Load(file.Provider("./config.json"), json.Parser()); err != nil {
		slog.Error("error loading config", err)
		return
	}
	var config models.NotionForwarderConfig
	if err := k.Unmarshal("", &config); err != nil {
		slog.Error("error loading config, could not unmarshal", err)
		return
	}

	client := dependencies.NewNotionForwardingClient(config.IntegrationToken, config.Databases)

	if err := client.PopulateForwardedDatabases(); err != nil {
		slog.Error("error populating database", err)
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

	slog.Info("Starting server on port :3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		return
	}
}
