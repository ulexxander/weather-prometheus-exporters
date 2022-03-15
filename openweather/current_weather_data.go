package openweather

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type CurrentWeatherDataResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  float64 `json:"pressure"`
		Humidity  float64 `json:"humidity"`
	} `json:"main"`
	Visibility float64 `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All float64 `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int     `json:"type"`
		ID      int     `json:"id"`
		Country string  `json:"country"`
		Sunrise float64 `json:"sunrise"`
		Sunset  float64 `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

type gauge struct {
	subsystem string
	name      string
	value     func(res *CurrentWeatherDataResponse) float64
	collector *prometheus.GaugeVec
}

type CurrentWeatherData struct {
	client   *Client
	lat      float64
	lon      float64
	interval time.Duration
	log      *log.Logger
	gauges   []gauge
}

func NewCurrentWeatherData(
	client *Client,
	lat, lon float64,
	interval time.Duration,
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
		client:   client,
		interval: interval,
		lat:      lat,
		lon:      lon,
		log:      log,
		gauges:   gauges,
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

func (cwd *CurrentWeatherData) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(cwd.interval):
			if err := cwd.Update(); err != nil {
				cwd.log.Println("Error updating current weather data:", err)
			}
		}
	}
}

func (cwd *CurrentWeatherData) Update() error {
	query := url.Values{}
	query.Set("lat", fmt.Sprint(cwd.lat))
	query.Set("lon", fmt.Sprint(cwd.lon))

	var res CurrentWeatherDataResponse
	if err := cwd.client.Request("/weather", query, &res); err != nil {
		return fmt.Errorf("requesting /weather: %w", err)
	}

	labels := prometheus.Labels{
		"id":   strconv.Itoa(res.ID),
		"name": res.Name,
	}

	for _, g := range cwd.gauges {
		val := g.value(&res)
		g.collector.With(labels).Set(val)
	}

	return nil
}
