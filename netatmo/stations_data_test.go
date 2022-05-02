package netatmo_test

import (
	"log"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"github.com/ulexxander/weather-prometheus-exporters/config"
	"github.com/ulexxander/weather-prometheus-exporters/netatmo"
	"github.com/ulexxander/weather-prometheus-exporters/testutil"
)

func TestStationsData(t *testing.T) {
	handler := testutil.NewHTTPHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	client := netatmo.NewClient(oauth)
	client.URL = server.URL

	stationsData := netatmo.NewStationsData(client, &config.NetatmoStationsData{}, log.Default())

	reg := prometheus.NewRegistry()
	err := reg.Register(stationsData)
	require.NoError(t, err)

	updated := make(chan struct{})
	go func() {
		stationsData.Update()
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

	indoorModuleMetric := func(name string, value float64) *dto.MetricFamily {
		return &dto.MetricFamily{
			Name: sptr("netatmo_indoor_module_" + name),
			Type: dto.MetricType_GAUGE.Enum(),
			Help: sptr(""),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{Name: sptr("home_id"), Value: sptr("61b646afb535277ce721d1a4")},
						{Name: sptr("home_name"), Value: sptr("My home")},
						{Name: sptr("id"), Value: sptr("70:ee:50:80:26:fa")},
						{Name: sptr("station_name"), Value: sptr("My home (Indoor)")},
						{Name: sptr("type"), Value: sptr("NAMain")},
					},
					Gauge: &dto.Gauge{
						Value: fptr(value),
					},
				},
			},
		}
	}

	outdoorModuleMetric := func(name string, value float64) *dto.MetricFamily {
		return &dto.MetricFamily{
			Name: sptr("netatmo_outdoor_module_" + name),
			Type: dto.MetricType_GAUGE.Enum(),
			Help: sptr(""),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{Name: sptr("home_id"), Value: sptr("61b646afb535277ce721d1a4")},
						{Name: sptr("home_name"), Value: sptr("My home")},
						{Name: sptr("id"), Value: sptr("02:00:00:7f:e6:96")},
						{Name: sptr("module_name"), Value: sptr("Zunanji modul")},
						{Name: sptr("type"), Value: sptr("NAModule1")},
					},
					Gauge: &dto.Gauge{
						Value: fptr(value),
					},
				},
			},
		}
	}

	windModuleMetric := func(name string, value float64) *dto.MetricFamily {
		return &dto.MetricFamily{
			Name: sptr("netatmo_wind_module_" + name),
			Type: dto.MetricType_GAUGE.Enum(),
			Help: sptr(""),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{Name: sptr("home_id"), Value: sptr("61b646afb535277ce721d1a4")},
						{Name: sptr("home_name"), Value: sptr("My home")},
						{Name: sptr("id"), Value: sptr("06:00:00:05:c6:48")},
						{Name: sptr("module_name"), Value: sptr("Veternica")},
						{Name: sptr("type"), Value: sptr("NAModule2")},
					},
					Gauge: &dto.Gauge{
						Value: fptr(value),
					},
				},
			},
		}
	}

	expectedMetrics := []*dto.MetricFamily{
		indoorModuleMetric("absolute_pressure", 965.5),
		indoorModuleMetric("co2", 762),
		indoorModuleMetric("humidity", 49),
		indoorModuleMetric("noise", 50),
		indoorModuleMetric("pressure", 1012),
		indoorModuleMetric("temperature", 20.9),
		outdoorModuleMetric("humidity", 91),
		outdoorModuleMetric("temperature", 11.9),
		windModuleMetric("gust_angle", 23),
		windModuleMetric("gust_strength", 5),
		windModuleMetric("wind_angle", 270),
		windModuleMetric("wind_strength", 1),
	}

	require.Equal(t, expectedMetrics, gatheredMetrics)
}

const response = `{
  "body": {
    "devices": [
      {
        "_id": "70:ee:50:80:26:fa",
        "date_setup": 1639335599,
        "last_setup": 1651169293,
        "type": "NAMain",
        "last_status_store": 1651477546,
        "firmware": 181,
        "wifi_status": 41,
        "reachable": true,
        "co2_calibrating": false,
        "data_type": ["Temperature", "CO2", "Humidity", "Noise", "Pressure"],
        "place": {
          "altitude": 395,
          "city": "Kranj",
          "country": "SI",
          "timezone": "Europe/Belgrade",
          "location": [14.3548565, 46.246113]
        },
        "station_name": "My home (Indoor)",
        "home_id": "61b646afb535277ce721d1a4",
        "home_name": "My home",
        "dashboard_data": {
          "time_utc": 1651477543,
          "Temperature": 20.9,
          "CO2": 762,
          "Humidity": 49,
          "Noise": 50,
          "Pressure": 1012,
          "AbsolutePressure": 965.5,
          "min_temp": 18.8,
          "max_temp": 28,
          "date_max_temp": 1651475120,
          "date_min_temp": 1651465946,
          "temp_trend": "down",
          "pressure_trend": "down"
        },
        "modules": [
          {
            "_id": "06:00:00:05:c6:48",
            "type": "NAModule2",
            "module_name": "Veternica",
            "last_setup": 1651333088,
            "data_type": ["Wind"],
            "battery_percent": 100,
            "reachable": true,
            "firmware": 25,
            "last_message": 1651477539,
            "last_seen": 1651477539,
            "rf_status": 74,
            "battery_vp": 6285,
            "dashboard_data": {
              "time_utc": 1651477539,
              "WindStrength": 1,
              "WindAngle": 270,
              "GustStrength": 5,
              "GustAngle": 23,
              "max_wind_str": 11,
              "max_wind_angle": 355,
              "date_max_wind_str": 1651467760
            }
          },
          {
            "_id": "02:00:00:7f:e6:96",
            "type": "NAModule1",
            "module_name": "Zunanji modul",
            "last_setup": 1651475006,
            "data_type": ["Temperature", "Humidity"],
            "battery_percent": 100,
            "reachable": true,
            "firmware": 50,
            "last_message": 1651477539,
            "last_seen": 1651477494,
            "rf_status": 87,
            "battery_vp": 6312,
            "dashboard_data": {
              "time_utc": 1651477494,
              "Temperature": 11.9,
              "Humidity": 91,
              "min_temp": 11.9,
              "max_temp": 20.8,
              "date_max_temp": 1651475085,
              "date_min_temp": 1651477494,
              "temp_trend": "down"
            }
          }
        ]
      }
    ],
    "user": {
      "mail": "socket.ttt.ttt@gmail.com",
      "administrative": {
        "lang": "en",
        "reg_locale": "en-US",
        "country": "SI",
        "unit": 0,
        "windunit": 0,
        "pressureunit": 2,
        "feel_like_algo": 0
      }
    }
  },
  "status": "ok",
  "time_exec": 0.08455610275268555,
  "time_server": 1651477647
}`
