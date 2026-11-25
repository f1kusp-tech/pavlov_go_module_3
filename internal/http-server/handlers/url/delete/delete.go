package delete

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	urlsvc "url-shortener/internal/service/url"
)

type Response struct {
	resp.Response
	CountDeleted int64 `json:"countDeleted"`
}

type URLDeleter interface {
	Delete(alias string) (int64, error)
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.delete.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("empty alias")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		countDeleted, err := deleter.Delete(alias)
		if errors.Is(err, urlsvc.ErrNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err), slog.String("alias", alias))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to delete url"))
			return
		}

		log.Info("deleted url", slog.String("alias", alias), slog.Int64("count_deleted", countDeleted))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response:     resp.OK(),
			CountDeleted: countDeleted,
		})
	}
}