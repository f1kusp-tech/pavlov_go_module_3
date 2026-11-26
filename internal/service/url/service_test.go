package url_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	urlsvc "url-shortener/internal/service/url"
)

func TestAliasLengthFromEnv(t *testing.T) {
	t.Parallel()

	n, err := urlsvc.AliasLengthFromEnv("")
	require.NoError(t, err)
	require.Equal(t, 5, n)

	_, err = urlsvc.AliasLengthFromEnv("0")
	require.Error(t, err)

	_, err = urlsvc.AliasLengthFromEnv("-1")
	require.Error(t, err)

	n, err = urlsvc.AliasLengthFromEnv("8")
	require.NoError(t, err)
	require.Equal(t, 8, n)
}