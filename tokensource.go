package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"os"
	"path/filepath"
)

type cachedTokenFile struct {
	path string
	src  oauth2.TokenSource
}

func newCachedTokenFile(ctx context.Context, bootstrap *oauth2.Token, path string) (oauth2.TokenSource, error) {
	var refreshSource oauth2.TokenSource
	if bootstrap != nil {
		if err := persistToken(path, bootstrap); err != nil {
			return nil, err
		}
		refreshSource = googleOAuth2.TokenSource(ctx, bootstrap)
	} else {
		tok, err := loadToken(path)
		if err != nil {
			return nil, fmt.Errorf("bootstrap token was not given, and failed to read local token: %w", err)
		}
		refreshSource = googleOAuth2.TokenSource(ctx, tok)
	}
	return oauth2.ReuseTokenSource(nil, cachedTokenFile{
		path: path,
		src:  refreshSource,
	}), nil
}

func (ct cachedTokenFile) Token() (*oauth2.Token, error) {
	token, _ := loadToken(ct.path)
	if token.Valid() {
		return token, nil
	}

	token, err := ct.src.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	if err := persistToken(ct.path, token); err != nil {
		return nil, fmt.Errorf("failed to persist token into file: %w", err)
	}
	return token, nil
}

func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var t *oauth2.Token
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return t, nil
}

func persistToken(path string, t *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(t); err != nil {
		return fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return nil
}
