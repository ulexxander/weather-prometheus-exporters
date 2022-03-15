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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulexxander/open-weather-prometheus-exporter/openweather"
)

var (
	flagAddr     = flag.String("addr", ":80", "Address to serve HTTP metrics on")
	flagAppID    = flag.String("app-id", "", "OpenWeather API application ID")
	flagCoords   = flag.String("coords", "", `Coordinates as JSON array with objects {"lat":123,"lon":123}`)
	flagInterval = flag.Duration("interval", time.Minute, "Interval how often data will be fetched and updated.")
)

func main() {
	log := log.Default()
	if err := run(log); err != nil {
		log.Fatalln("Fatal error:", err)
	}
}

type coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func run(log *log.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	flag.Parse()
	if *flagAppID == "" {
		return fmt.Errorf("flag app-id is missing")
	}
	if *flagCoords == "" {
		return fmt.Errorf("flag coords is missing")
	}

	var coords []coord
	if err := json.Unmarshal([]byte(*flagCoords), &coords); err != nil {
		return fmt.Errorf("unmarshaling coordinates: %w", err)
	}
	if len(coords) == 0 {
		return fmt.Errorf("no coordinates specified")
	}

	log.Println("Starting OpenWeather Prometheus Exporter")

	client := openweather.NewClient(*flagAppID)

	jobErrors := make(chan error, len(coords))

	for _, c := range coords {
		go func(c coord) {
			cwd := openweather.NewCurrentWeatherData(client, c.Lat, c.Lon, *flagInterval, log)
			if err := prometheus.Register(cwd); err != nil {
				jobErrors <- fmt.Errorf("registering collector: %w", err)
				return
			}

			log.Printf("Starting update job for lat=%f lon=%f", c.Lat, c.Lon)
			jobErrors <- cwd.Run(ctx)
		}(c)
	}

	httpErr := make(chan error)
	go func() {
		log.Println("Starting HTTP server on", *flagAddr)
		if err := http.ListenAndServe(*flagAddr, promhttp.Handler()); err != nil {
			httpErr <- err
		}
	}()

	select {
	case err := <-httpErr:
		return fmt.Errorf("listening HTTP: %w", err)
	case err := <-jobErrors:
		return fmt.Errorf("job errors: %w", err)
	}
}
