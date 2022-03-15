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
