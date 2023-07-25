package viridian

import (
	"context"
	"errors"
)

type loginRequest struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

type loginResponse struct {
	Token        string `json:"token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func Login(ctx context.Context, key, secret string) (API, error) {
	var api API
	if key == "" {
		return api, errors.New("api key cannot be blank")
	}
	if secret == "" {
		return api, errors.New("api secret cannot be blank")
	}
	c := loginRequest{
		APIKey:    key,
		APISecret: secret,
	}
	resp, err := doPost[loginRequest, loginResponse](ctx, "/customers/api/login", api, c)
	if err != nil {
		return api, err
	}
	api.Token = resp.Token
	api.ExpiresIn = resp.ExpiresIn
	api.RefreshToken = resp.RefreshToken
	return api, nil
}
