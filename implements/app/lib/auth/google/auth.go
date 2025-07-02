package google

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthenticatedUser struct {
	Subject  string
	Email    string
	Issuer   string
	Name     string
	ImageURL *string
}

var (
	p *oidc.Provider
	c *oauth2.Config
)

func Init() error {
	provider, err := oidc.NewProvider(context.Background(), os.Getenv("PROVIDER"))
	if err != nil {
		return errors.Wrap(err, "failed to provice google")
	}

	p = provider

	c = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return nil
}

func GetAuthURL() string {
	return c.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func GetAuthenticatedUser(ctx context.Context, authCode string) (*AuthenticatedUser, *time.Time, error) {
	oauth2Token, err := c.Exchange(ctx, authCode)

	if err != nil {
		return nil, nil, err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)

	if !ok {
		return nil, nil, err
	}

	idToken, err := p.Verifier(&oidc.Config{ClientID: os.Getenv("CLIENT_ID")}).Verify(ctx, rawIDToken)

	if err != nil {
		return nil, nil, err
	}

	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		return nil, nil, err
	}

	email, ok := profile["email"].(string)

	if !ok {
		return nil, nil, fmt.Errorf("invalid email")
	}

	issuer, ok := profile["iss"].(string)

	if !ok {
		return nil, nil, fmt.Errorf("invalid iss")
	}

	name, ok := profile["name"].(string)

	if !ok {
		return nil, nil, fmt.Errorf("invalid name")
	}

	var imageURL *string
	picture, ok := profile["picture"].(string)

	if ok {
		imageURL = &picture
	}

	exp, ok := profile["exp"].(float64)

	if !ok {
		return nil, nil, fmt.Errorf("invalid exp")
	}

	expire := time.Unix(int64(exp), 0)

	return &AuthenticatedUser{
		Subject:  idToken.Subject,
		Email:    email,
		Issuer:   issuer,
		Name:     name,
		ImageURL: imageURL,
	}, &expire, nil
}
