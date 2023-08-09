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

func Login(ctx context.Context, secretPrefix, key, secret, apiBase string) (API, error) {
	if key == "" {
		return API{}, errors.New("api key cannot be blank")
	}
	if secret == "" {
		return API{}, errors.New("api secret cannot be blank")
	}
	c := loginRequest{
		APIKey:    key,
		APISecret: secret,
	}
	resp, err := doPost[loginRequest, loginResponse](ctx, apiBase+"/customers/api/login", "", c)
	if err != nil {
		return API{}, err
	}
	api := NewAPI(secretPrefix, key, secret, resp.Token, apiBase)
	return *api, nil
}
