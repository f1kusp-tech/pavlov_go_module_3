package update

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	urlsvc "url-shortener/internal/service/url"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLUpdater interface {
	Update(alias, urlToSave string) error
}

func New(log *slog.Logger, updater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.update.New"

		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request body"))
			return
		}

		err := updater.Update(alias, req.URL)
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}
		if errors.Is(err, urlsvc.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to update url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to update url"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}