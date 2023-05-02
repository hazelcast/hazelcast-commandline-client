package viridian

import (
	"context"
	"errors"
)

type loginRequest struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

type legacyLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func Login(ctx context.Context, key, secret string) (API, error) {
	if LegacyAPI() {
		return loginLegacy(ctx, key, secret)
	}
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
	api.token = resp.Token
	return api, nil
}

func loginLegacy(ctx context.Context, email, password string) (API, error) {
	var api API
	if email == "" {
		return api, errors.New("email cannot be blank")
	}
	if password == "" {
		return api, errors.New("password cannot be blank")
	}
	c := legacyLoginRequest{
		Email:    email,
		Password: password,
	}
	resp, err := doPost[legacyLoginRequest, loginResponse](ctx, "/customers/login", "", c)
	if err != nil {
		return api, err
	}
	api.token = resp.Token
	return api, nil
}
