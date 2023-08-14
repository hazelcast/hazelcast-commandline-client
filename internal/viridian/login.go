package viridian

import (
	"context"
)

type loginRequest struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func Login(ctx context.Context, secretPrefix, key, secret, apiBase string) (API, error) {
	a := API{
		SecretPrefix: secretPrefix,
		Key:          key,
		Secret:       secret,
		APIBaseURL:   apiBase,
	}
	if err := a.login(ctx); err != nil {
		return a, err
	}
	return a, nil
}
