package openweather_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ulexxander/open-weather-prometheus-exporter/openweather"
	"github.com/ulexxander/open-weather-prometheus-exporter/testutil"
)

func TestClient_Error(t *testing.T) {
	handler := testutil.NewHTTPHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	client := openweather.NewClient("my-app-id")
	client.URL = server.URL

	type requestResult struct {
		cwd *openweather.CurrentWeatherDataResponse
		err error
	}
	resultChan := make(chan requestResult)
	go func() {
		cwd, err := client.CurrentWeatherData(openweather.Coordinates{
			Lat: 46.2389,
			Lon: 14.3556,
		})
		resultChan <- requestResult{cwd, err}
	}()

	var r *http.Request
	select {
	case r = <-handler.Requests:
	case <-time.After(time.Second):
		require.Fail(t, "request did not arrived")
	}

	require.Equal(t, "/weather", r.URL.Path)

	expectedQuery := url.Values{}
	expectedQuery.Set("appid", "my-app-id")
	expectedQuery.Set("lat", "46.2389")
	expectedQuery.Set("lon", "14.3556")
	require.Equal(t, expectedQuery, r.URL.Query())

	response := openweather.ErrorResponse{
		Cod:     123,
		Message: "something went wrong",
	}
	handler.Responses <- response

	var result requestResult
	select {
	case result = <-resultChan:
	case <-time.After(time.Second):
		require.Fail(t, "result did not arrived")
	}

	require.Equal(t, &response, result.err)
	require.Nil(t, result.cwd)
}
