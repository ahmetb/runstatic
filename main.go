package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"log"
)

func main() {
	ctx := context.Background()

	// ensureAuthenticated
	ts, err := tokenSource(ctx)
	if err != nil {
		log.Fatalf("authentication error: %v", err)
	}
	_ = ts
	oauthSvc, err := googleoauth2.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatal(err)
	}

	// ensure can read email
	tokInfo, err := oauthSvc.Tokeninfo().Do()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully logged in as " + color.GreenString(tokInfo.Email) + "!")

	// ensure default project
	projectID, err := ensureDefaultProject(ctx, ts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Default " + gcp + " project is set to " + color.New(color.FgGreen, color.Bold).Sprint(projectID) + ".")
}

func ensureDefaultProject(ctx context.Context, ts oauth2.TokenSource) (string, error) {
	p, ok, err := defaultProject()
	if err != nil {
		return "", err
	}
	if !ok {
		v, err := chooseProject(ctx, ts)
		if err != nil {
			return "", err
		}
		p = v.id
		if err := setDefaultProject(p); err != nil {
			return "", fmt.Errorf("failed to save default project: %w", err)
		}
	}
	return p, nil
}

// tokenSource prompts for login if not logged in, otherwise returns the existing token source.
func tokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	ok, err := credsFile.Exists()
	if err != nil {
		return nil, err
	}
	var bootstrapToken *oauth2.Token
	if !ok {
		tok, err := authenticate()
		if err != nil {
			return nil, err
		}
		bootstrapToken = tok
	}
	return newCachedTokenFile(ctx, bootstrapToken, credsFile.Path())
}
