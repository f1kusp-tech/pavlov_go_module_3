package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	urlsvc "url-shortener/internal/service/url"
)

func TestSave_Success(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	urlSaverMock.On("Save", "https://google.com", "test_alias").
		Return("test_alias", nil).
		Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(`{"url":"https://google.com","alias":"test_alias"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp save.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "OK", resp.Status)
	require.Equal(t, "test_alias", resp.Alias)
}

func TestSave_BadRequest(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(`{invalid json`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp save.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "Error", resp.Status)
	require.NotEmpty(t, resp.Error)
}

func TestSave_Conflict(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	urlSaverMock.On("Save", "https://google.com", "busy").
		Return("", urlsvc.ErrAlreadyExists).
		Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(`{"url":"https://google.com","alias":"busy"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)

	var resp save.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "Error", resp.Status)
}

func TestSave_InternalServerError(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	urlSaverMock.On("Save", "https://google.com", "ok").
		Return("", errors.New("database unavailable")).
		Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(`{"url":"https://google.com","alias":"ok"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSave_BadRequestInvalidAlias(t *testing.T) {
	svc := urlsvc.New(newPanickingStorage(t), 5)
	handler := save.New(slogdiscard.NewDiscardLogger(), svc)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(`{"url":"https://google.com","alias":"bad/slash"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

type panickingStorage struct {
	t *testing.T
}

func newPanickingStorage(t *testing.T) *panickingStorage {
	t.Helper()
	return &panickingStorage{t: t}
}

func (f *panickingStorage) SaveURL(string, string) (int64, error) {
	f.t.Fatal("storage must not be called for invalid alias")
	return 0, nil
}

func (f *panickingStorage) GetURL(string) (string, error)        { return "", nil }
func (f *panickingStorage) DeleteURL(string) (int64, error)      { return 0, nil }
func (f *panickingStorage) UpdateURL(string, string) error       { return nil }
func (f *panickingStorage) AliasExists(string) (bool, error)   { return false, nil }
