package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"urlShortener/iternal/config"
	"urlShortener/iternal/http-server/handlers/url/get"
	save "urlShortener/iternal/http-server/handlers/url/save"
	logger "urlShortener/iternal/http-server/middleware"
	"urlShortener/iternal/pkg/logger/sl"
	"urlShortener/iternal/storage/mon"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	Local = "local"
	Dev   = "dev"
	Prod  = "prod"
)

func main() {
	//config load
	cfg := config.MustLoad()

	//logger load
	log := loggerSet(cfg.Env)

	log.Info("Starting app", slog.String("env", cfg.Env))
	log.Debug("Debug messages are enabled")

	//storage load
	storage, err := mon.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	_ = storage

	//router set
	router := chi.NewRouter()

	//middlwares
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//handlers
	router.Post("/url", save.New(log, cfg, storage))
	router.Get("/url/{alias}", get.New(log, cfg, storage))

	log.Info("starting server", slog.String("address", cfg.ServerSet.Adress))

	srv := &http.Server{
		Addr:         cfg.ServerSet.Adress,
		Handler:      router,
		ReadTimeout:  cfg.ServerSet.Timeout,
		WriteTimeout: cfg.ServerSet.Timeout,
		IdleTimeout:  cfg.ServerSet.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	storage.Client.Disconnect(context.TODO())

	log.Info("Server stopped")
}

func loggerSet(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case Local:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case Dev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case Prod:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
