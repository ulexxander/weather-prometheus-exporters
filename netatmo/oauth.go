package netatmo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type OAuth interface {
	Token(scope string) (*OAuthTokenResponse, error)
}

type oauth struct {
	URL          string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

const DefaultOAuthURL = "https://api.netatmo.com/oauth2"

func NewOAuth(clientID, clientSecret, username, password string) *oauth {
	return &oauth{
		URL:          DefaultOAuthURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
	}
}

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Token obtains access token using client credentials grant type.
// Docs: This method can only be used with the same account that the one who owns the API application.
func (oa *oauth) Token(scope string) (*OAuthTokenResponse, error) {
	var res OAuthTokenResponse
	body := url.Values{}
	body.Set("grant_type", "password")
	body.Set("client_id", oa.ClientID)
	body.Set("client_secret", oa.ClientSecret)
	body.Set("username", oa.Username)
	body.Set("password", oa.Password)
	body.Set("scope", scope)
	if err := oa.Request("/token", body, &res); err != nil {
		return nil, fmt.Errorf("requesting /token: %w", err)
	}
	return &res, nil
}

func (oa *oauth) Request(endpoint string, reqBody url.Values, dest interface{}) error {
	var body io.Reader
	if reqBody != nil {
		urlEncoded := reqBody.Encode()
		body = strings.NewReader(urlEncoded)
	}

	url := oa.URL + endpoint
	res, err := http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return fmt.Errorf("sending HTTP POST request: %w", err)
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
