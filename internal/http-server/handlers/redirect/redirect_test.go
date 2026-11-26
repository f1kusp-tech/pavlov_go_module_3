package redirect_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	urlsvc "url-shortener/internal/service/url"
)

type panickingStorage struct {
	t *testing.T
}

func newPanickingStorage(t *testing.T) *panickingStorage {
	t.Helper()
	return &panickingStorage{t: t}
}

func (f *panickingStorage) SaveURL(string, string) (int64, error) {
	f.t.Fatal("storage must not be called")
	return 0, nil
}

func (f *panickingStorage) GetURL(string) (string, error)        { return "", nil }
func (f *panickingStorage) DeleteURL(string) (int64, error)      { return 0, nil }
func (f *panickingStorage) UpdateURL(string, string) error       { return nil }
func (f *panickingStorage) AliasExists(string) (bool, error)     { return false, nil }

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

func TestRedirect_BadRequestInvalidAlias(t *testing.T) {
	svc := urlsvc.New(newPanickingStorage(t), 5)
	handler := redirect.New(slogdiscard.NewDiscardLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/bad/slash", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "bad/slash")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRedirect_InternalServerError(t *testing.T) {
	handler := redirect.New(slogdiscard.NewDiscardLogger(), &mockResolver{err: errors.New("database unavailable")})

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}