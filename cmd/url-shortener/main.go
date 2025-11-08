package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	//ssogrpc "github.com/YaroslavalsoraY/TrainingProject/internal/clients/sso/grpc"
	"github.com/YaroslavalsoraY/TrainingProject/internal/config"
	"github.com/YaroslavalsoraY/TrainingProject/internal/http-server/handlers/url/delete"
	"github.com/YaroslavalsoraY/TrainingProject/internal/http-server/handlers/url/redirect"
	"github.com/YaroslavalsoraY/TrainingProject/internal/http-server/handlers/url/save"
	"github.com/YaroslavalsoraY/TrainingProject/internal/http-server/middleware/logger"
	"github.com/YaroslavalsoraY/TrainingProject/internal/lib/logger/handlers/slogpretty"
	"github.com/YaroslavalsoraY/TrainingProject/internal/lib/logger/sl"
	eventsender "github.com/YaroslavalsoraY/TrainingProject/internal/services/event-sender"
	"github.com/YaroslavalsoraY/TrainingProject/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// ssoClient, err := ssogrpc.New(
	// 	context.Background(),
	// 	log,
	// 	cfg.Clients.SSO.Adress,
	// 	cfg.Clients.SSO.Timeout,
	// 	cfg.Clients.SSO.RetriesCount,
	// )

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	// ssoClient.IsAdmin(context.Background(), 1)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})
	
	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("adress", cfg.Adress))

	srv := &http.Server{
		Addr: cfg.Adress,
		Handler: router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}
	
	sender := eventsender.New(storage, log)
	sender.StartProcessingEvents(context.Background(), 5 * time.Second)

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

	// TODO: run server: 
}

func setupLogger(env string) *slog.Logger{
	var log *slog.Logger
	
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}