package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"golang.org/x/oauth2"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/option"
	"os"
)

// ensureBilling holds the user until they set up billing on Cloud Console.
func ensureBilling(ctx context.Context, ts oauth2.TokenSource, projectID string) error {
	client, err := cloudbilling.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("failed to initialize cloud billing client: %w", err)
	}

	// wait until billing is enabled on the project
	for {
		if ok, err := projectBillingEnabled(ctx, client.Projects, projectID); err != nil {
			return fmt.Errorf("could not query billing information for project %q: %w", projectID, err)
		} else if ok {
			return nil
		}

		accounts, err := listBillingAccounts(ctx, client.BillingAccounts)
		if err != nil {
			return fmt.Errorf("failed to list cloud billing accounts: %w", err)
		}

		if len(accounts) == 0 {
			// user has no billing account, likely new GCP user
			fmt.Println("Looks like it's your " + color.GreenString("first time") + " here!")
			fmt.Println("We need to " + color.New(color.Bold, color.FgYellow).Sprint("set up billing!"))
			fmt.Println("But don't worry, Google Cloud has a generous free tier:\n  " +
				color.New(color.FgCyan, color.Underline).Sprint("https://cloud.google.com/free"))
			fmt.Println("Go here, and set up billing for your project:")
			fmt.Println("  " + color.New(color.FgHiBlue, color.Underline).Sprintf("https://console.cloud.google.com/billing/linkedaccount?project=%s", projectID))
			fmt.Printf(color.HiBlackString("When done, hit \"Return\" to continue: "))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		} else {
			// pick an existing billing account directly here
			fmt.Println("We need to " + color.New(color.Bold, color.FgYellow).Sprint("set up billing for this project."))

			choice, err := promptForAccounts(accounts)
			if err != nil {
				return err
			}

			if err := setProjectBillingAccount(ctx, client.Projects, projectID, choice); err != nil {
				return fmt.Errorf("failed to update billing account for project %q: %w", projectID, err)
			}

		}
	}
}

func promptForAccounts(accounts []*cloudbilling.BillingAccount) (*cloudbilling.BillingAccount, error) {
	var options []string
	optionMap := make(map[string]*cloudbilling.BillingAccount)

	for _, a := range accounts {
		name := a.DisplayName
		if a.DisplayName == "" {
			name = a.Name
		}
		options = append(options, name)
		optionMap[name] = a
	}

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: "Choose a billing method for this project:",
		Options: options,
	}, &choice, survey.WithValidator(survey.Required))
	return optionMap[choice], err
}

func setProjectBillingAccount(ctx context.Context, client *cloudbilling.ProjectsService, projectID string, account *cloudbilling.BillingAccount) error {
	_, err := client.UpdateBillingInfo("projects/"+projectID, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: account.Name,
	}).Context(ctx).Do()
	return err
}

func projectBillingEnabled(ctx context.Context, client *cloudbilling.ProjectsService, projectID string) (bool, error) {
	bo, err := client.GetBillingInfo("projects/" + projectID).Context(ctx).Do()
	if err != nil {
		return false, fmt.Errorf("failed to query project billing info: %w", err)
	}
	return bo.BillingEnabled, nil
}

func listBillingAccounts(ctx context.Context, client *cloudbilling.BillingAccountsService) ([]*cloudbilling.BillingAccount, error) {
	var out []*cloudbilling.BillingAccount
	var pageToken string
	for {
		resp, err := client.List().Context(ctx).PageToken(pageToken).Do()
		if err != nil {
			return nil, err
		}

		// add only open accounts
		for _, v := range resp.BillingAccounts {
			if v.Open {
				out = append(out, v)
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return out, nil
}
