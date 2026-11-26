package tests

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/update"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const host = "localhost:8082"

func loadAuthFromEnv(t *testing.T) (user, password string) {
	t.Helper()

	for _, path := range []string{"../.env", ".env", "../.env.example", ".env.example"} {
		_ = godotenv.Load(path)
	}

	user = os.Getenv("AUTH_USER")
	password = os.Getenv("AUTH_PASSWORD")
	if user == "" || password == "" {
		t.Fatal("AUTH_USER and AUTH_PASSWORD must be set (.env or .env.example)")
	}

	return user, password
}

func requireServer(t *testing.T) *httpexpect.Expect {
	t.Helper()

	conn, err := net.DialTimeout("tcp", host, time.Second)
	if err != nil {
		t.Skipf("server not running at %s: %v", host, err)
	}
	_ = conn.Close()

	u := url.URL{Scheme: "http", Host: host}
	return httpexpect.Default(t, u.String())
}

func TestURLShortener_Unauthorized(t *testing.T) {
	e := requireServer(t)

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(6),
		}).
		Expect().
		Status(http.StatusUnauthorized)

	e.PUT("/url/some-alias").
		WithJSON(update.Request{URL: gofakeit.URL()}).
		Expect().
		Status(http.StatusUnauthorized)

	e.DELETE("/url/some-alias").
		Expect().
		Status(http.StatusUnauthorized)
}

func TestURLShortener_HappyPath(t *testing.T) {
	user, password := loadAuthFromEnv(t)
	e := requireServer(t)

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth(user, password).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirectRemove(t *testing.T) {
	user, password := loadAuthFromEnv(t)

	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field url is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := requireServer(t)

			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth(user, password).
				Expect()

			if tc.error != "" {
				resp.Status(http.StatusBadRequest).
					JSON().Object().
					Value("error").String().IsEqual(tc.error)
				return
			}

			obj := resp.Status(http.StatusCreated).JSON().Object()

			alias := tc.alias
			if alias != "" {
				obj.Value("alias").String().IsEqual(tc.alias)
			} else {
				obj.Value("alias").String().NotEmpty()
				alias = obj.Value("alias").String().Raw()
			}

			testRedirect(t, alias, tc.url)

			e.DELETE("/url/"+alias).
				WithBasicAuth(user, password).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("countDeleted").Number().Gt(0)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	t.Helper()

	u := url.URL{Scheme: "http", Host: host, Path: alias}
	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectedToURL)
}