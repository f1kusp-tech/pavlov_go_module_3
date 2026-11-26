package delete_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/delete"
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

type mockDeleter struct {
	count int64
	err   error
}

func (m *mockDeleter) Delete(alias string) (int64, error) {
	return m.count, m.err
}

func TestDelete_Success(t *testing.T) {
	handler := delete.New(slogdiscard.NewDiscardLogger(), &mockDeleter{count: 1})

	req := httptest.NewRequest(http.MethodDelete, "/url/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp delete.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, int64(1), resp.CountDeleted)
}

func TestDelete_NotFound(t *testing.T) {
	handler := delete.New(slogdiscard.NewDiscardLogger(), &mockDeleter{err: urlsvc.ErrNotFound})

	req := httptest.NewRequest(http.MethodDelete, "/url/missing", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDelete_BadRequestInvalidAlias(t *testing.T) {
	svc := urlsvc.New(newPanickingStorage(t), 5)
	handler := delete.New(slogdiscard.NewDiscardLogger(), svc)

	req := httptest.NewRequest(http.MethodDelete, "/url/bad/slash", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "bad/slash")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDelete_InternalServerError(t *testing.T) {
	handler := delete.New(slogdiscard.NewDiscardLogger(), &mockDeleter{err: errors.New("database unavailable")})

	req := httptest.NewRequest(http.MethodDelete, "/url/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}