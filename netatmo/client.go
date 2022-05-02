package netatmo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	URL   string
	OAuth OAuth
}

const DefaultURL = "https://api.netatmo.com/api"

func NewClient(oauth OAuth) *Client {
	return &Client{
		URL:   DefaultURL,
		OAuth: oauth,
	}
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("netatmo: code=%d msg=%s", e.Code, e.Message)
}

type ErrorResponse struct {
	Error *Error `json:"error"`
}

type StationsDataResponse struct {
	ErrorResponse
	Body struct {
		Devices []Device `json:"devices"`
		User    struct {
			Mail           string `json:"mail"`
			Administrative struct {
				Lang         string `json:"lang"`
				RegLocale    string `json:"reg_locale"`
				Country      string `json:"country"`
				Unit         int    `json:"unit"`
				Windunit     int    `json:"windunit"`
				Pressureunit int    `json:"pressureunit"`
				FeelLikeAlgo int    `json:"feel_like_algo"`
			} `json:"administrative"`
		} `json:"user"`
	} `json:"body"`
	Status     string  `json:"status"`
	TimeExec   float64 `json:"time_exec"`
	TimeServer int     `json:"time_server"`
}

const (
	DeviceTypeIndoor  = "NAMain"
	DeviceTypeOutdoor = "NAModule1"
	DeviceTypeWind    = "NAModule2"
)

type Device struct {
	ID              string   `json:"_id"`
	DateSetup       int      `json:"date_setup"`
	LastSetup       int      `json:"last_setup"`
	Type            string   `json:"type"`
	LastStatusStore int      `json:"last_status_store"`
	Firmware        int      `json:"firmware"`
	WifiStatus      int      `json:"wifi_status"`
	Reachable       bool     `json:"reachable"`
	Co2Calibrating  bool     `json:"co2_calibrating"`
	DataType        []string `json:"data_type"`
	Place           struct {
		Altitude int       `json:"altitude"`
		City     string    `json:"city"`
		Country  string    `json:"country"`
		Timezone string    `json:"timezone"`
		Location []float64 `json:"location"`
	} `json:"place"`
	StationName   string `json:"station_name"`
	HomeID        string `json:"home_id"`
	HomeName      string `json:"home_name"`
	DashboardData struct {
		TimeUtc int `json:"time_utc"`
		IndoorModuleData
	} `json:"dashboard_data"`
	Modules []Module `json:"modules"`
}

type Module struct {
	ID             string   `json:"_id"`
	Type           string   `json:"type"`
	ModuleName     string   `json:"module_name"`
	LastSetup      int      `json:"last_setup"`
	DataType       []string `json:"data_type"`
	BatteryPercent int      `json:"battery_percent"`
	Reachable      bool     `json:"reachable"`
	Firmware       int      `json:"firmware"`
	LastMessage    int      `json:"last_message"`
	LastSeen       int      `json:"last_seen"`
	RfStatus       int      `json:"rf_status"`
	BatteryVp      int      `json:"battery_vp"`
	DashboardData  struct {
		TimeUtc int `json:"time_utc"`
		OutdoorModuleData
		WindModuleData
	} `json:"dashboard_data"`
}

type IndoorModuleData struct {
	Temperature      float64 `json:"Temperature"`
	CO2              float64 `json:"CO2"`
	Humidity         float64 `json:"Humidity"`
	Noise            float64 `json:"Noise"`
	Pressure         float64 `json:"Pressure"`
	AbsolutePressure float64 `json:"AbsolutePressure"`
	MinTemp          float64 `json:"min_temp"`
	MaxTemp          float64 `json:"max_temp"`
	DateMaxTemp      int     `json:"date_max_temp"`
	DateMinTemp      int     `json:"date_min_temp"`
	PressureTrend    string  `json:"pressure_trend"`
}

type OutdoorModuleData struct {
	Temperature float64 `json:"Temperature"`
	Humidity    float64 `json:"Humidity"`
	MinTemp     float64 `json:"min_temp"`
	MaxTemp     float64 `json:"max_temp"`
	DateMaxTemp int     `json:"date_max_temp"`
	DateMinTemp int     `json:"date_min_temp"`
	TempTrend   string  `json:"temp_trend"`
}

type WindModuleData struct {
	WindStrength   float64 `json:"WindStrength"`
	WindAngle      float64 `json:"WindAngle"`
	GustStrength   float64 `json:"GustStrength"`
	GustAngle      float64 `json:"GustAngle"`
	MaxWindStr     float64 `json:"max_wind_str"`
	MaxWindAngle   float64 `json:"max_wind_angle"`
	DateMaxWindStr int     `json:"date_max_wind_str"`
}

func (c *Client) StationsData() (*StationsDataResponse, error) {
	var res StationsDataResponse
	if err := c.Request("/getstationsdata", &res); err != nil {
		return nil, fmt.Errorf("requesting /getstationsdata: %w", err)
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &res, nil
}

func (c *Client) Request(endpoint string, dest interface{}) error {
	token, err := c.OAuth.Token("read_station")
	if err != nil {
		return fmt.Errorf("obtaining OAuth access token: %w", err)
	}

	url := c.URL + endpoint
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("initializing HTTP request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending HTTP GET request: %w", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(resBody, dest); err != nil {
		return fmt.Errorf("unmarshaling response body: %w, content: %s", err, resBody)
	}

	return nil
}
