package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"notion-forwarder/dependencies"
	_ "notion-forwarder/docs"
	"notion-forwarder/handlers"
	"notion-forwarder/models"
)

// @title Notion Forwarder API
// @version 1.0
// @description This is the API for the Notion Forwarder service

// @license.name AGPL-3.0
// @license.url https://www.gnu.org/licenses/agpl-3.0.html

// @BasePath /

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

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:3000/swagger/doc.json"),
	))

	slog.Info("Starting server on port :3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		return
	}
}
