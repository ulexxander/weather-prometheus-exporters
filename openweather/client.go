package openweather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultURL = "https://api.openweathermap.org/data/2.5"

type Client struct {
	URL   string
	AppID string
}

func NewClient(appID string) *Client {
	return &Client{
		URL:   DefaultURL,
		AppID: appID,
	}
}

type ErrorResponse struct {
	Cod     int    `json:"cod"`
	Message string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("openweather: cod=%d message=%s", e.Cod, e.Message)
}

func (e *ErrorResponse) OK() bool {
	return e.Message == ""
}

type CurrentWeatherDataResponse struct {
	ErrorResponse
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
}

func (c *Client) CurrentWeatherData(lat, lon float64) (*CurrentWeatherDataResponse, error) {
	query := url.Values{}
	query.Set("lat", fmt.Sprint(lat))
	query.Set("lon", fmt.Sprint(lon))

	var res CurrentWeatherDataResponse
	if err := c.Request("/weather", query, &res); err != nil {
		return nil, fmt.Errorf("requesting /weather: %w", err)
	}

	if !res.OK() {
		return nil, &res.ErrorResponse
	}

	return &res, nil
}

func (c *Client) Request(endpoint string, query url.Values, dest interface{}) error {
	if query == nil {
		query = url.Values{}
	}
	query.Add("appid", c.AppID)

	url := c.URL + endpoint + "?" + query.Encode()
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("sending HTTP GET request: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("unmarshaling response body: %w, content: %s", err, body)
	}

	return nil
}
