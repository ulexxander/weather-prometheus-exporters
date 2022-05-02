package netatmo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ulexxander/open-weather-prometheus-exporter/netatmo"
	"github.com/ulexxander/open-weather-prometheus-exporter/testutil"
)

type oauthMock struct {
	res *netatmo.OAuthTokenResponse
}

func (oa *oauthMock) Token(scope string) (*netatmo.OAuthTokenResponse, error) {
	return oa.res, nil
}

var oauth = &oauthMock{
	res: &netatmo.OAuthTokenResponse{
		AccessToken:  "f1894h0quehf",
		ExpiresIn:    3600,
		RefreshToken: "1uu01huiefffa",
	},
}

func TestClient_Error(t *testing.T) {
	handler := testutil.NewHTTPHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	client := netatmo.NewClient(oauth)
	client.URL = server.URL

	type requestResult struct {
		stationsData *netatmo.StationsDataResponse
		err          error
	}
	resultChan := make(chan requestResult)
	go func() {
		stationsData, err := client.StationsData()
		resultChan <- requestResult{stationsData, err}
	}()

	var r *http.Request
	select {
	case r = <-handler.Requests:
	case <-time.After(time.Second):
		require.Fail(t, "request did not arrived")
	}

	require.Equal(t, "/getstationsdata", r.URL.Path)
	require.Equal(t, r.Header.Get("Authorization"), "Bearer f1894h0quehf")

	response := netatmo.ErrorResponse{
		Error: &netatmo.Error{
			Code:    123,
			Message: "something went wrong",
		},
	}
	handler.Responses <- response

	var result requestResult
	select {
	case result = <-resultChan:
	case <-time.After(time.Second):
		require.Fail(t, "result did not arrived")
	}

	require.Equal(t, response.Error, result.err)
	require.Nil(t, result.stationsData)
}
