package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"url-shortener/internal/lib/logger/sl"
	urlsvc "url-shortener/internal/service/url"
)

type URLResolver interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, resolver URLResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.redirect.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			http.NotFound(w, r)
			return
		}

		url, err := resolver.GetURL(alias)
		if errors.Is(err, urlsvc.ErrNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err), slog.String("alias", alias))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}