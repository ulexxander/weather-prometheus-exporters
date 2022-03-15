package openweather_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ulexxander/open-weather-prometheus-exporter/openweather"
)

func TestCurrentWeatherData(t *testing.T) {
	appID, ok := os.LookupEnv("OPEN_WEATHER_APP_ID")
	if !ok {
		t.Fatalf("OPEN_WEATHER_APP_ID env variable must be set")
	}

	client := openweather.NewClient(appID)
	kranjLat, kranjLon := 46.2389, 14.3556
	interval := time.Millisecond * 100
	cwd := openweather.NewCurrentWeatherData(client, kranjLat, kranjLon, interval, log.Default())

	reg := prometheus.NewRegistry()
	if err := reg.Register(cwd); err != nil {
		t.Fatalf("error registering collector: %v", err)
	}

	if err := cwd.Update(); err != nil {
		t.Fatalf("error updating: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("error gathering metrics: %v", err)
	}
	if len(mfs) != 1 {
		t.Fatalf("expected 1 metric family, got: %d", len(mfs))
	}

	for _, mf := range mfs {
		metrics := mf.GetMetric()
		if len(metrics) != 1 {
			t.Fatalf("expected 1 metric, got: %d", len(metrics))
		}
		gauge := metrics[0].GetGauge()
		if *gauge.Value == 0 {
			t.Fatalf("gauge value is zero")
		}
	}
}
