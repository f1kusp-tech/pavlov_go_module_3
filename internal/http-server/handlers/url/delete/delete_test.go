package delete_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	urlsvc "url-shortener/internal/service/url"
)

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