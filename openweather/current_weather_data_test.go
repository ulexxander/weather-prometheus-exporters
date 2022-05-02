package openweather_test

import (
	"log"
	"net/http/httptest"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/ulexxander/weather-prometheus-exporters/config"
	"github.com/ulexxander/weather-prometheus-exporters/openweather"
	"github.com/ulexxander/weather-prometheus-exporters/testutil"
)

func TestCurrentWeatherData(t *testing.T) {
	handler := testutil.NewHTTPHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	client := openweather.NewClient("my-app-id")
	client.URL = server.URL

	config := &config.OpenWeatherCurrentWeatherData{
		Coords: []config.Coordinates{
			{
				Lat: 46.2389,
				Lon: 14.3556,
			},
		},
	}
	cwd := openweather.NewCurrentWeatherData(client, config, log.Default())

	reg := prometheus.NewRegistry()
	err := reg.Register(cwd)
	require.NoError(t, err)

	updated := make(chan struct{})
	go func() {
		cwd.Update()
		updated <- struct{}{}
	}()

	select {
	case <-handler.Requests:
		handler.Responses <- []byte(response)
	case <-time.After(time.Second):
		require.Fail(t, "request did not arrived")
	}

	<-updated

	gatheredMetrics, err := reg.Gather()
	require.NoError(t, err)

	sptr := func(s string) *string { return &s }
	fptr := func(f float64) *float64 { return &f }

	metric := func(name string, value float64) *dto.MetricFamily {
		return &dto.MetricFamily{
			Name: sptr("open_weather_" + name),
			Type: dto.MetricType_GAUGE.Enum(),
			Help: sptr(""),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{Name: sptr("id"), Value: sptr("3197378")},
						{Name: sptr("name"), Value: sptr("Kranj")},
					},
					Gauge: &dto.Gauge{
						Value: fptr(value),
					},
				},
			},
		}
	}

	expectedMetrics := []*dto.MetricFamily{
		metric("clouds_all", 75),
		metric("main_feels_like", 287.29),
		metric("main_humidity", 72),
		metric("main_pressure", 1015),
		metric("main_temp", 287.88),
		metric("main_temp_max", 289.04),
		metric("main_temp_min", 284.16),
		metric("wind_deg", 290),
		metric("wind_speed", 3.6),
	}

	require.Equal(t, expectedMetrics, gatheredMetrics)
}

const response = `{
  "coord": { "lon": 14.3556, "lat": 46.2389 },
  "weather": [
    {
      "id": 803,
      "main": "Clouds",
      "description": "broken clouds",
      "icon": "04d"
    }
  ],
  "base": "stations",
  "main": {
    "temp": 287.88,
    "feels_like": 287.29,
    "temp_min": 284.16,
    "temp_max": 289.04,
    "pressure": 1015,
    "humidity": 72
  },
  "visibility": 10000,
  "wind": { "speed": 3.6, "deg": 290 },
  "clouds": { "all": 75 },
  "dt": 1651487420,
  "sys": {
    "type": 1,
    "id": 6815,
    "country": "SI",
    "sunrise": 1651463259,
    "sunset": 1651515081
  },
  "timezone": 7200,
  "id": 3197378,
  "name": "Kranj",
  "cod": 200
}`
