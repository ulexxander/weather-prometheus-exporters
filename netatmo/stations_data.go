package netatmo

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type StationsData struct {
	client              *Client
	log                 *log.Logger
	indoorModuleGauges  []indoorModuleGauge
	outdoorModuleGauges []outdoorModuleGauge
	windModuleGauges    []windModuleGauge
}

type indoorModuleGauge struct {
	name      string
	value     func(data *IndoorModuleData) float64
	collector *prometheus.GaugeVec
}

type outdoorModuleGauge struct {
	name      string
	value     func(data *OutdoorModuleData) float64
	collector *prometheus.GaugeVec
}

type windModuleGauge struct {
	name      string
	value     func(data *WindModuleData) float64
	collector *prometheus.GaugeVec
}

func NewStationsData(client *Client, log *log.Logger) *StationsData {
	const namespace = "netatmo"
	stationLabels := []string{"home_id", "home_name", "id", "type", "station_name"}
	moduleLabels := []string{"home_id", "home_name", "id", "type", "module_name"}

	indoorModuleGauges := []indoorModuleGauge{
		{
			name:  "absolute_pressure",
			value: func(data *IndoorModuleData) float64 { return data.AbsolutePressure },
		},
		{
			name:  "co2",
			value: func(data *IndoorModuleData) float64 { return data.CO2 },
		},
		{
			name:  "humidity",
			value: func(data *IndoorModuleData) float64 { return data.Humidity },
		},
		{
			name:  "noise",
			value: func(data *IndoorModuleData) float64 { return data.Noise },
		},
		{
			name:  "pressure",
			value: func(data *IndoorModuleData) float64 { return data.Pressure },
		},
		{
			name:  "temperature",
			value: func(data *IndoorModuleData) float64 { return data.Temperature },
		},
	}

	outdoorModuleGauges := []outdoorModuleGauge{
		{
			name:  "humidity",
			value: func(data *OutdoorModuleData) float64 { return data.Humidity },
		},
		{
			name:  "temperature",
			value: func(data *OutdoorModuleData) float64 { return data.Temperature },
		},
	}

	windModuleGauges := []windModuleGauge{
		{
			name:  "gust_angle",
			value: func(data *WindModuleData) float64 { return data.GustAngle },
		},
		{
			name:  "gust_strength",
			value: func(data *WindModuleData) float64 { return data.GustStrength },
		},
		{
			name:  "wind_angle",
			value: func(data *WindModuleData) float64 { return data.WindAngle },
		},
		{
			name:  "wind_strength",
			value: func(data *WindModuleData) float64 { return data.WindStrength },
		},
	}

	for i := range indoorModuleGauges {
		g := &indoorModuleGauges[i]
		g.collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "indoor_module",
			Name:      g.name,
		}, stationLabels)
	}

	for i := range outdoorModuleGauges {
		g := &outdoorModuleGauges[i]
		g.collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "outdoor_module",
			Name:      g.name,
		}, moduleLabels)
	}

	for i := range windModuleGauges {
		g := &windModuleGauges[i]
		g.collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "wind_module",
			Name:      g.name,
		}, moduleLabels)
	}

	return &StationsData{
		client:              client,
		log:                 log,
		indoorModuleGauges:  indoorModuleGauges,
		outdoorModuleGauges: outdoorModuleGauges,
		windModuleGauges:    windModuleGauges,
	}
}

func (sd *StationsData) forEach(f func(c prometheus.Collector)) {
	for _, g := range sd.indoorModuleGauges {
		f(g.collector)
	}
	for _, g := range sd.outdoorModuleGauges {
		f(g.collector)
	}
	for _, g := range sd.windModuleGauges {
		f(g.collector)
	}
}

func (sd *StationsData) Describe(d chan<- *prometheus.Desc) {
	sd.forEach(func(c prometheus.Collector) {
		c.Describe(d)
	})
}

func (sd *StationsData) Collect(m chan<- prometheus.Metric) {
	sd.forEach(func(c prometheus.Collector) {
		c.Collect(m)
	})
}

func (sd *StationsData) Update() {
	start := time.Now()

	sd.log.Print("Fetching stations data")

	stationsData, err := sd.client.StationsData()
	if err != nil {
		sd.log.Printf("Error fetching stations data: %s", err)
		return
	}

	for _, device := range stationsData.Body.Devices {
		stationLabels := prometheus.Labels{
			"home_id":      device.HomeID,
			"home_name":    device.HomeName,
			"id":           device.ID,
			"type":         device.Type,
			"station_name": device.StationName,
		}

		for _, g := range sd.indoorModuleGauges {
			val := g.value(&device.DashboardData.IndoorModuleData)
			g.collector.With(stationLabels).Set(val)
		}

		sd.log.Printf("Processed dashboard data of %s device %s (%s)", device.Type, device.StationName, device.ID)

		for _, module := range device.Modules {
			moduleLabels := prometheus.Labels{
				"home_id":     device.HomeID,
				"home_name":   device.HomeName,
				"id":          module.ID,
				"type":        module.Type,
				"module_name": module.ModuleName,
			}

			switch module.Type {
			case DeviceTypeOutdoor:
				for _, g := range sd.outdoorModuleGauges {
					val := g.value(&module.DashboardData.OutdoorModuleData)
					g.collector.With(moduleLabels).Set(val)
				}
			case DeviceTypeWind:
				for _, g := range sd.windModuleGauges {
					val := g.value(&module.DashboardData.WindModuleData)
					g.collector.With(moduleLabels).Set(val)
				}
			default:
				sd.log.Printf("Unsupported module type: %s", module.Type)
			}

			sd.log.Printf("Processed dashboard data of %s module %s (%s)", module.Type, module.ModuleName, module.ID)
		}
	}

	duration := time.Since(start)
	sd.log.Println("Updated stations data successfully, took", duration)
}
