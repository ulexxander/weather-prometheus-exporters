package openweather_test

import (
	"log"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/ulexxander/open-weather-prometheus-exporter/openweather"
)

func TestCurrentWeatherData(t *testing.T) {
	appID := os.Getenv("OPEN_WEATHER_APP_ID")
	if appID == "" {
		require.Fail(t, "OPEN_WEATHER_APP_ID env variable must be set")
	}

	client := openweather.NewClient(appID)
	config := &openweather.CurrentWeatherDataConfig{
		Coords: []openweather.Coordinates{
			// Kranj.
			{
				Lat: 46.2389,
				Lon: 14.3556,
			},
		},
		Interval: openweather.Duration(time.Millisecond * 100),
	}
	cwd := openweather.NewCurrentWeatherData(client, config, log.Default())

	reg := prometheus.NewRegistry()
	err := reg.Register(cwd)
	require.NoError(t, err)

	cwd.Update()

	expectedMetrics := []string{
		"open_weather_clouds_all",
		"open_weather_main_feels_like",
		"open_weather_main_humidity",
		"open_weather_main_pressure",
		"open_weather_main_temp",
		"open_weather_main_temp_max",
		"open_weather_main_temp_min",
		"open_weather_wind_deg",
		"open_weather_wind_speed",
	}

	mfs, err := reg.Gather()
	require.NoError(t, err)

	var gatheredMetrics []string
	for _, mf := range mfs {
		gatheredMetrics = append(gatheredMetrics, mf.GetName())
	}
	sort.Strings(gatheredMetrics)
	require.Equal(t, expectedMetrics, gatheredMetrics)

	for _, mf := range mfs {
		metrics := mf.GetMetric()
		require.Len(t, metrics, 1)

		firstMetric := metrics[0]
		labels := firstMetric.GetLabel()
		require.Len(t, labels, 2)

		expectedLabels := []string{"id", "name"}
		var labelNames []string
		for _, l := range labels {
			labelNames = append(labelNames, l.GetName())
		}
		require.Equal(t, expectedLabels, labelNames)

		gauge := firstMetric.GetGauge()
		require.NotNil(t, gauge.Value)
	}
}
