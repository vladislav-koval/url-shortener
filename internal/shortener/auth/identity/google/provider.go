// Package google — драйвер для Google OAuth/OIDC, реализует service.IdentityProvider.
// Все обращения к SDK Google (обмен кода, валидация id_token) собраны здесь,
// а не в транспорте — это внешняя технология, а не HTTP-специфика запроса.
package google

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"

	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

type Provider struct {
	oauthConfig *oauth2.Config
}

func NewProvider(cfg Config) *Provider {
	googleOAuth := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleCallbackURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     googleoauth.Endpoint,
	}

	return &Provider{oauthConfig: googleOAuth}
}

func (p *Provider) Exchange(ctx context.Context, code string, verifier string) (domain.GoogleIdentity, error) {
	token, err := p.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return domain.GoogleIdentity{}, fmt.Errorf("exchange authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return domain.GoogleIdentity{}, fmt.Errorf("google token response missing id_token")
	}

	payload, err := idtoken.Validate(ctx, rawIDToken, p.oauthConfig.ClientID)
	if err != nil {
		return domain.GoogleIdentity{}, fmt.Errorf("validate google id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)

	return domain.GoogleIdentity{
		Sub:           payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
	}, nil
}

func (p *Provider) AuthCodeURL(state string, verifier string) string {
	return p.oauthConfig.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(verifier),
	)
}
