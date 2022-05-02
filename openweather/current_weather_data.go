package openweather

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ulexxander/open-weather-prometheus-exporter/config"
)

type gauge struct {
	subsystem string
	name      string
	value     func(res *CurrentWeatherDataResponse) float64
	collector *prometheus.GaugeVec
}

type CurrentWeatherData struct {
	client *Client
	config *config.OpenWeatherCurrentWeatherData
	log    *log.Logger
	gauges []gauge
}

func NewCurrentWeatherData(
	client *Client,
	config *config.OpenWeatherCurrentWeatherData,
	log *log.Logger,
) *CurrentWeatherData {
	const namespace = "open_weather"
	labels := []string{"id", "name"}

	gauges := []gauge{
		// Main gauges.
		{
			subsystem: "main",
			name:      "temp",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Main.Temp },
		},
		{
			subsystem: "main",
			name:      "feels_like",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Main.FeelsLike },
		},
		{
			subsystem: "main",
			name:      "temp_min",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Main.TempMin },
		},
		{
			subsystem: "main",
			name:      "temp_max",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Main.TempMax },
		},
		{
			subsystem: "main",
			name:      "pressure",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Main.Pressure },
		},
		{
			subsystem: "main",
			name:      "humidity",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Main.Humidity },
		},
		// Wind gauges.
		{
			subsystem: "wind",
			name:      "speed",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Wind.Speed },
		},
		{
			subsystem: "wind",
			name:      "deg",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Wind.Deg },
		},
		// Clouds gauges.
		{
			subsystem: "clouds",
			name:      "all",
			value:     func(res *CurrentWeatherDataResponse) float64 { return res.Clouds.All },
		},
	}

	for i := range gauges {
		g := &gauges[i]
		g.collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: g.subsystem,
			Name:      g.name,
		}, labels)
	}

	return &CurrentWeatherData{
		client: client,
		config: config,
		log:    log,
		gauges: gauges,
	}
}

func (cwd *CurrentWeatherData) Describe(d chan<- *prometheus.Desc) {
	for _, g := range cwd.gauges {
		g.collector.Describe(d)
	}
}

func (cwd *CurrentWeatherData) Collect(m chan<- prometheus.Metric) {
	for _, g := range cwd.gauges {
		g.collector.Collect(m)
	}
}

func (cwd *CurrentWeatherData) Run(ctx context.Context) {
	cwd.Update()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(cwd.config.Interval)):
			cwd.Update()
		}
	}
}

func (cwd *CurrentWeatherData) Update() {
	type result struct {
		res *CurrentWeatherDataResponse
		err error
	}

	results := make(chan result, len(cwd.config.Coords))
	start := time.Now()

	for _, coords := range cwd.config.Coords {
		go func(coords config.Coordinates) {
			res, err := cwd.client.CurrentWeatherData(coords.Lat, coords.Lon)
			results <- result{res, err}
		}(coords)
	}

	for i := 0; i < len(cwd.config.Coords); i++ {
		result := <-results
		if result.err != nil {
			cwd.log.Println("Error fetching Current Weather Data:", result.err)
			continue
		}

		labels := prometheus.Labels{
			"id":   strconv.Itoa(result.res.ID),
			"name": result.res.Name,
		}

		for _, g := range cwd.gauges {
			val := g.value(result.res)
			g.collector.With(labels).Set(val)
		}

		cwd.log.Printf("Processed Current Weather Data of %s (%d)", result.res.Name, result.res.ID)
	}

	duration := time.Since(start)
	cwd.log.Println("Updated Current Weather Data successfully, took", duration)
}
