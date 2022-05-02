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
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulexxander/weather-prometheus-exporters/config"
	"github.com/ulexxander/weather-prometheus-exporters/netatmo"
	"github.com/ulexxander/weather-prometheus-exporters/openweather"
)

var (
	flagAddr    = flag.String("addr", ":80", "Address to serve HTTP metrics on")
	flagConfig  = flag.String("config", "./config.json", "Config file location")
	flagEnvFile = flag.String("env-file", "", "Environment variables file to load (dotenv)")
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

	log.Print("Parsing flags")
	flag.Parse()

	if *flagEnvFile != "" {
		log.Println("Loading environment variables from", *flagEnvFile)
		if err := godotenv.Load(*flagEnvFile); err != nil {
			return fmt.Errorf("loading .env file: %w", err)
		}
	}

	log.Println("Reading config file from", *flagConfig)
	configJSON, err := os.ReadFile(*flagConfig)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	var config config.Config
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := runOpenWeather(ctx, &config.OpenWeather, log); err != nil {
		return fmt.Errorf("running OpenWeather: %w", err)
	}
	if err := runNetatmo(ctx, &config.Netatmo, log); err != nil {
		return fmt.Errorf("running Netatmo: %w", err)
	}

	log.Println("Starting HTTP server on", *flagAddr)
	server := http.Server{
		Addr:    *flagAddr,
		Handler: promhttp.Handler(),
	}

	shutdownDone := make(chan error)
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		log.Print("Shutting down HTTP server")
		shutdownDone <- server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("serving HTTP: %w", err)
	}

	<-shutdownDone

	return nil
}

type env struct {
	missingKeys []string
}

func (e *env) Get(key string) string {
	val := os.Getenv(key)
	if val == "" {
		e.missingKeys = append(e.missingKeys, key)
	}
	return val
}

func (e *env) Error() error {
	if len(e.missingKeys) == 0 {
		return nil
	}
	missing := strings.Join(e.missingKeys, ", ")
	return fmt.Errorf("missing environment variables: %s", missing)
}

func runOpenWeather(ctx context.Context, config *config.OpenWeather, log *log.Logger) error {
	if !config.CurrentWeatherData.Enabled {
		log.Print("OpenWeather Current Weather Data is disabled")
		return nil
	}

	var env env
	appID := env.Get("OPEN_WEATHER_APP_ID")
	if err := env.Error(); err != nil {
		return err
	}

	client := openweather.NewClient(appID)

	cwd := openweather.NewCurrentWeatherData(client, &config.CurrentWeatherData, log)
	if err := prometheus.Register(cwd); err != nil {
		return fmt.Errorf("registering Current Weather Data collector: %w", err)
	}

	log.Print("Starting OpenWeather Current Weather Data job")
	go cwd.Run(ctx)

	return nil
}

func runNetatmo(ctx context.Context, config *config.Netatmo, log *log.Logger) error {
	if !config.StationsData.Enabled {
		log.Print("Netatmo Stations Data is disabled")
		return nil
	}

	var env env
	clientID := env.Get("NETATMO_CLIENT_ID")
	clientSecret := env.Get("NETATMO_CLIENT_SECRET")
	username := env.Get("NETATMO_USERNAME")
	password := env.Get("NETATMO_PASSWORD")
	if err := env.Error(); err != nil {
		return err
	}

	oauth := netatmo.NewOAuth(clientID, clientSecret, username, password)
	client := netatmo.NewClient(oauth)

	stationsData := netatmo.NewStationsData(client, &config.StationsData, log)
	if err := prometheus.Register(stationsData); err != nil {
		return fmt.Errorf("registering Stations Data collector: %w", err)
	}

	log.Print("Starting Stations Data update job")
	go stationsData.Run(ctx)

	return nil
}
