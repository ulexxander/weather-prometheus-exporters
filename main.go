package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulexxander/open-weather-prometheus-exporter/openweather"
)

var (
	flagAddr   = flag.String("addr", ":80", "Address to serve HTTP metrics on")
	flagConfig = flag.String("config", "./config.json", "Config file location")
)

func main() {
	log := log.Default()
	if err := run(log); err != nil {
		log.Fatalln("Fatal error:", err)
	}
}

func run(log *log.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	flag.Parse()
	appID := os.Getenv("OPEN_WEATHER_APP_ID")
	if appID == "" {
		return fmt.Errorf("OPEN_WEATHER_APP_ID env variable must be set")
	}

	log.Println("Reading config file from", *flagConfig)

	configJSON, err := os.ReadFile(*flagConfig)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	var config openweather.Config
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	client := openweather.NewClient(appID)

	cwd := openweather.NewCurrentWeatherData(client, &config.CurrentWeatherData, log)
	if err := prometheus.Register(cwd); err != nil {
		return fmt.Errorf("registering current weather data collector: %w", err)
	}

	cwdErr := make(chan error)
	go func() {
		log.Print("Starting current weather data update job")
		cwdErr <- cwd.Run(ctx)
	}()

	httpErr := make(chan error)
	go func() {
		log.Println("Starting HTTP server on", *flagAddr)
		httpErr <- http.ListenAndServe(*flagAddr, promhttp.Handler())
	}()

	select {
	case err := <-cwdErr:
		return fmt.Errorf("current weather data: %w", err)
	case err := <-httpErr:
		return fmt.Errorf("HTTP server: %w", err)
	}
}
