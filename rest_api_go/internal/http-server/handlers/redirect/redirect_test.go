package redirect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	urlsvc "url-shortener/internal/service/url"
)

type mockResolver struct {
	url string
	err error
}

func (m *mockResolver) GetURL(alias string) (string, error) {
	return m.url, m.err
}

func TestRedirect_Success(t *testing.T) {
	handler := redirect.New(slogdiscard.NewDiscardLogger(), &mockResolver{url: "https://example.com"})

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusFound, rr.Code)
	require.Equal(t, "https://example.com", rr.Header().Get("Location"))
}

func TestRedirect_NotFound(t *testing.T) {
	handler := redirect.New(slogdiscard.NewDiscardLogger(), &mockResolver{err: urlsvc.ErrNotFound})

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}