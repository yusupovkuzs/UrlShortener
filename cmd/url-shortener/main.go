package main

import (
	// project
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/http-server/handlers/save"
	mwLogger "go-url-shortener/internal/http-server/middleware/logger"
	"go-url-shortener/internal/lib/logger/handlers/slogpretty"
	"go-url-shortener/internal/lib/logger/sl"
	"go-url-shortener/internal/storage/postgres"
	"go-url-shortener/internal/http-server/handlers/redirect"
	"go-url-shortener/internal/http-server/handlers/delete"

	// embedded
	"log/slog"
	"os"
	"net/http"

	// external
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init config: cleanenv
	cfg := config.MustLoad()

	// init logger: slog
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// init storage: postgress
	storage, err := postgres.NewStorage(cfg.Postgres)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	_ = storage

	// init router: chi, "chi render"
	router := chi.NewRouter()
	// middleware
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	// autherization
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HttpServer.User: cfg.HttpServer.Password,
		}))	

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})
	// handlers
	router.Get("/{alias}", redirect.New(log, storage)) 

	// start server
	log.Info("starting server", slog.String("address", cfg.HttpServer.Address))
	srv := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}
	
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to stop server")
		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
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
