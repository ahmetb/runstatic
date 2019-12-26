package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"golang.org/x/oauth2"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

type project struct {
	number      int64
	id          string
	displayName string
}

func projectsClient(ctx context.Context, ts oauth2.TokenSource) (*cloudresourcemanager.Service, error) {
	s, err := cloudresourcemanager.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudresourcemanager client: %w", err)
	}
	return s, nil
}

func listProjects(ctx context.Context, client *cloudresourcemanager.Service) ([]project, error) {
	var out []project
	var pageToken string
	for {
		resp, err := client.Projects.List().Context(ctx).PageSize(1000).PageToken(pageToken).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list GCP projects: %w", err)
		}
		for _, p := range resp.Projects {
			out = append(out, project{
				number:      p.ProjectNumber,
				id:          p.ProjectId,
				displayName: p.Name,
			})
		}
		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return out, nil
}

func chooseProject(ctx context.Context, ts oauth2.TokenSource) (*project, error) {
	client, err := projectsClient(ctx, ts)
	if err != nil {
		return nil, err
	}

	list, err := listProjects(ctx, client)
	if err != nil {
		return nil, err
	}
	list = nil

	if len(list) == 0 {
		fmt.Println(color.RedString("You don't have any ") + gcp + color.RedString(" projects.") + " Let's create one!")
		fmt.Println(color.HiBlackString("(Projects lets you separate your cloud usage from other things.)"))
		// Prompt for a new project to be created.
		for {
			requestedID, err := promptNewProjectID()
			if err != nil {
				return nil, fmt.Errorf("failed to prompt for new project ID: %w", err)
			}
			p, err := createProject(ctx, client, requestedID)
			if err != nil {
				fmt.Printf("ERROR: failed to create "+gcp+" project: %v\n", err)
				continue // until we get one
			}
			return p, nil
		}
	}

	bold := color.New(color.Bold)
	fmt.Printf("You have %s "+gcp+" projects. Let's make one your %s!\n", bold.Sprintf("%d", len(list)), bold.Sprint("default"))
	return promptExistingProject(list)
}

func promptNewProjectID() (string, error) {
	var out string
	err := survey.AskOne(&survey.Input{
		Message: "Choose a unique name for your new " + gcp + " project:",
		Help:    "Your project name is not visible to your users, though it cannot be changed later. It just helps you keep things organized.",
	}, &out, survey.WithValidator(survey.Required))
	// TODO(ahmetb): can add client-side project name validation here
	return out, err
}

func createProject(ctx context.Context, client *cloudresourcemanager.Service, id string) (*project, error) {
	op, err := client.Projects.Create(&cloudresourcemanager.Project{
		ProjectId: id,
		Name:      id, // use same Display Name as Project ID
		Labels: map[string]string{
			"created-via": "runstatic"},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	for !op.Done {
		op, err = client.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to query operation status (%s): %w", op.Name, err)
		}
	}
	if op.Error != nil {
		return nil, fmt.Errorf("create project failed (code=%d): %s", op.Error.Code, op.Error.Message)
	}

	var p cloudresourcemanager.Project
	if err := json.Unmarshal(op.Response, &p); err != nil {
		return nil, fmt.Errorf("failed to decode operation response as Project object: %w", err)
	}

	return &project{
		number:      p.ProjectNumber,
		id:          p.ProjectId,
		displayName: p.Name,
	}, nil
}

func promptExistingProject(list []project) (*project, error) {
	var options []string
	var optMap = make(map[string]project)
	for _, p := range list {
		displayName := p.id
		if p.displayName != "" && p.displayName != p.id {
			displayName += " (" + p.displayName + ")"
		}
		options = append(options, displayName)
		optMap[displayName] = p
	}

	var choice string
	if err := survey.AskOne(&survey.Select{
		Message: "Choose a project to deploy into:",
		Options: options,
	}, &choice,
		survey.WithValidator(survey.Required)); err != nil {
		return nil, fmt.Errorf("could not choose a project: %+v", err)
	}
	v := optMap[choice]
	return &v, nil
}
