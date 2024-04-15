package main

import (
	"errors"
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
	"strings"
)

// @title Notion Forwarder API
// @version 1.0
// @description This is the API for the Notion Forwarder service

// @license.name AGPL-3.0
// @license.url https://www.gnu.org/licenses/agpl-3.0.html

// @BasePath /

var k = koanf.New(".")

func parseConfig() (models.NotionForwarderConfig, error) {
	// Load config
	if err := k.Load(file.Provider("./config.json"), json.Parser()); err != nil {
		slog.Error("error loading config", err)
		return models.NotionForwarderConfig{}, errors.New("could not load config")
	}
	var config models.NotionForwarderConfig
	if err := k.Unmarshal("", &config); err != nil {
		slog.Error("error loading config, could not unmarshal", err)
		return models.NotionForwarderConfig{}, errors.New("could not unmarshal config")
	}
	return config, nil
}

func setLogLevel(logLevel string) error {
	level := strings.ToLower(logLevel)
	switch level {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
		break
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
		break
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
		break
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
		break
	default:
		return errors.New("invalid log level")
	}
	slog.Info("Set log level", "log_level", logLevel)
	return nil
}

func main() {
	config, err := parseConfig()
	if err != nil {
		slog.Error("error parsing config, exiting", err)
		return
	}

	if err := setLogLevel(config.LogLevel); err != nil {
		slog.Error("error setting log level", err)
		return
	}

	client := dependencies.NewNotionForwardingClient(config.IntegrationToken, config.Databases)

	if !config.LazyLoad {
		if err := client.PopulateForwardedDatabases(); err != nil {
			slog.Error("error populating database", err)
			return
		}
	} else {
		slog.Info("Lazy loading enabled, entries will be fetched on demand")
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
