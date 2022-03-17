package openweather_test

import (
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ulexxander/open-weather-prometheus-exporter/openweather"
)

func TestCurrentWeatherData(t *testing.T) {
	appID := os.Getenv("OPEN_WEATHER_APP_ID")
	if appID == "" {
		t.Fatalf("OPEN_WEATHER_APP_ID env variable must be set")
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
	if err := reg.Register(cwd); err != nil {
		t.Fatalf("error registering collector: %v", err)
	}

	if err := cwd.Update(); err != nil {
		t.Fatalf("error updating: %v", err)
	}

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
	if err != nil {
		t.Fatalf("error gathering metrics: %v", err)
	}

	var gatheredMetrics []string
	for _, mf := range mfs {
		gatheredMetrics = append(gatheredMetrics, mf.GetName())
	}
	sort.Strings(gatheredMetrics)
	if !reflect.DeepEqual(gatheredMetrics, expectedMetrics) {
		t.Fatalf("expected metrics %v, got: %v", expectedMetrics, gatheredMetrics)
	}

	for _, mf := range mfs {
		metrics := mf.GetMetric()
		if len(metrics) != 1 {
			t.Fatalf("%s: expected 1 metric in family, got: %d", mf.GetName(), len(metrics))
		}
		firstMetric := metrics[0]
		labels := firstMetric.GetLabel()
		if len(labels) != 2 {
			t.Fatalf("%s: expected 2 labels, got: %d", mf.GetName(), len(labels))
		}
		expectedLabels := []string{"id", "name"}
		var labelNames []string
		for _, l := range labels {
			labelNames = append(labelNames, l.GetName())
		}
		if !reflect.DeepEqual(labelNames, expectedLabels) {
			t.Fatalf("%s: expected labels %v, got: %v", mf.GetName(), expectedLabels, labelNames)
		}
		gauge := firstMetric.GetGauge()
		if gauge.Value == nil {
			t.Fatalf("%s: gauge value is nil", mf.GetName())
		}
	}
}
