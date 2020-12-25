package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	googleoauth2 "google.golang.org/api/oauth2/v2"
)

var (
	credsFile = mustDotDir().File("credentials.json")
)

var (
	googel = color.HiBlueString("G") +
		color.HiRedString("o") +
		color.YellowString("o") +
		color.HiBlueString("g") +
		color.HiGreenString("l") +
		color.HiRedString("e")
	gcp = googel + color.HiWhiteString(" Cloud")

	// from Google Developer Console (https://console.developers.google.com).
	googleOAuth2 = &oauth2.Config{
		// ClientID:     "233694408259-m4tvj6abiqr5dqtdh97sjvqs4mdp3o40.apps.googleusercontent.com",
		// ClientSecret: "V2mIxRWOcb6CnDsMXoUhXHKw",

		ClientID:     "32555940559.apps.googleusercontent.com",
		ClientSecret: "ZmssLNjJy2998hD4CTg2ejr2",
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	scopes = []string{
		googleoauth2.UserinfoEmailScope,         // required to read email
		cloudresourcemanager.CloudPlatformScope, // required for everything GCP
		// "https://www.googleapis.com/auth/cloudplatformprojects", // required to create projects, probably superfluous after cloud-platform scope.
	}
)

func authenticate() (*oauth2.Token, error) {
	url := googleOAuth2.AuthCodeURL("state")

	fmt.Println("runstatic uses your " + googel + " account to deploy.")
	fmt.Println(color.HiBlackString("Your credentials are not collected, and don't leave your machine."))
	fmt.Println("Let's authenticate to your " + googel + " account:")
	urlLabel := color.New(color.FgHiBlack, color.Underline).Sprint(url)
	open, err := openBrowser(url)
	if err == nil && open {
		fmt.Println(
			"  1. I just opened a browser for you, authorize there.\n" +
				"     (If the browser window did not come up, try this URL:\n" +
				"     " + urlLabel + " ).")
	} else {
		fmt.Println(
			"  1. Visit this URL and authorize the application:\n" +
				"     " + urlLabel)
	}
	fmt.Printf("\n" +
		"  2. " + color.New(color.Bold).Sprint("Copy") + " the code from browser, " +
		color.New(color.Bold).Sprint("paste here") + ": ")

	in, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}
	tok, err := googleOAuth2.Exchange(context.TODO(), in)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code for token: %w", err)
	}
	return tok, nil
}

func openBrowser(url string) (bool, error) {
	switch runtime.GOOS {
	case "linux":
		return true, exec.Command("xdg-open", url).Start()
	case "windows":
		return true, exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return true, exec.Command("open", url).Start()
	default:
		return false, nil
	}
}
