package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/update"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	urlsvc "url-shortener/internal/service/url"
	"url-shortener/internal/storage/postgres"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting the project...", slog.String("env", cfg.Env), slog.String("version", "v1"))

	storage, err := postgres.New(cfg.Database.DSN())
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	aliasLength, err := urlsvc.AliasLengthFromEnv(os.Getenv("ALIAS_LENGTH"))
	if err != nil {
		log.Error("failed to parse ALIAS_LENGTH", sl.Err(err))
		os.Exit(1)
	}
	log.Debug("alias length", slog.Int("length", aliasLength))

	urlService := urlsvc.New(storage, aliasLength)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.Auth.User: cfg.Auth.Password,
		}))

		r.Post("/", save.New(log, urlService))
		r.Put("/{alias}", update.New(log, urlService))
		r.Delete("/{alias}", delete.New(log, urlService))
	})

	router.Get("/{alias}", redirect.New(log, urlService))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to start server", sl.Err(err))
		os.Exit(1)
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return setupPrettySlog()
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
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