package netatmo_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ulexxander/open-weather-prometheus-exporter/netatmo"
	"github.com/ulexxander/open-weather-prometheus-exporter/testutil"
)

func TestOAuth(t *testing.T) {
	handler := testutil.NewHTTPHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	oauth := netatmo.NewOAuth(
		"my-clientID",
		"my-clientSecret",
		"my-username",
		"my-password",
	)
	oauth.URL = server.URL

	type requestResult struct {
		token *netatmo.OAuthTokenResponse
		err   error
	}
	resultChan := make(chan requestResult)
	go func() {
		token, err := oauth.Token("my-scope")
		resultChan <- requestResult{token, err}
	}()

	var r *http.Request
	select {
	case r = <-handler.Requests:
	case <-time.After(time.Second):
		require.Fail(t, "request did not arrived")
	}

	require.Equal(t, "/token", r.URL.Path)

	ct := r.Header.Get("Content-Type")
	require.Equal(t, "application/x-www-form-urlencoded", ct)

	bodyRaw, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	body, err := url.ParseQuery(string(bodyRaw))
	require.NoError(t, err)

	expectedBody := url.Values{}
	expectedBody.Set("grant_type", "password")
	expectedBody.Set("client_id", "my-clientID")
	expectedBody.Set("client_secret", "my-clientSecret")
	expectedBody.Set("username", "my-username")
	expectedBody.Set("password", "my-password")
	expectedBody.Set("scope", "my-scope")
	require.Equal(t, expectedBody, body)

	response := netatmo.OAuthTokenResponse{
		AccessToken:  "i2c34r3480rc8n02yu34uhf",
		ExpiresIn:    3600,
		RefreshToken: "2423u-9fc-8y2y8-9fy",
	}
	handler.Responses <- response

	var result requestResult
	select {
	case result = <-resultChan:
	case <-time.After(time.Second):
		require.Fail(t, "result did not arrived")
	}

	require.NoError(t, result.err)
	require.Equal(t, response, *result.token)
}
