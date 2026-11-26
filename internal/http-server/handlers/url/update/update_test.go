package update_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/update"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	urlsvc "url-shortener/internal/service/url"
)

type mockUpdater struct {
	err error
}

func (m *mockUpdater) Update(alias, urlToSave string) error {
	return m.err
}

func TestUpdate_Success(t *testing.T) {
	handler := update.New(slogdiscard.NewDiscardLogger(), &mockUpdater{})

	req := httptest.NewRequest(http.MethodPut, "/url/abc", bytes.NewReader([]byte(`{"url":"https://example.com"}`)))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp update.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "abc", resp.Alias)
}

func TestUpdate_BadRequest(t *testing.T) {
	handler := update.New(slogdiscard.NewDiscardLogger(), &mockUpdater{})

	req := httptest.NewRequest(http.MethodPut, "/url/abc", bytes.NewReader([]byte(`not json`)))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdate_NotFound(t *testing.T) {
	handler := update.New(slogdiscard.NewDiscardLogger(), &mockUpdater{err: urlsvc.ErrNotFound})

	req := httptest.NewRequest(http.MethodPut, "/url/missing", bytes.NewReader([]byte(`{"url":"https://example.com"}`)))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdate_InternalServerError(t *testing.T) {
	handler := update.New(slogdiscard.NewDiscardLogger(), &mockUpdater{err: errors.New("database unavailable")})

	req := httptest.NewRequest(http.MethodPut, "/url/abc", bytes.NewReader([]byte(`{"url":"https://example.com"}`)))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}