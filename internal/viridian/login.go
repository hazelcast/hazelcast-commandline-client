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
	Token string `json:"token"`
}

func Login(ctx context.Context, secretPrefix, key, secret string) (API, error) {
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
	resp, err := doPost[loginRequest, loginResponse](ctx, "/customers/api/login", "", c)
	if err != nil {
		return api, err
	}
	api.key = key
	api.secret = secret
	api.token = resp.Token
	api.secretPrefix = secretPrefix
	return api, nil
}
