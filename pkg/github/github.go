package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// Client represents a GitHub client
type Client struct {
	client *github.Client
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{client: client}, nil
}

// Authenticate verifies the GitHub token by fetching the authenticated user
func (c *Client) Authenticate() error {
	ctx := context.Background()
	user, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}
	fmt.Printf("Authenticated as GitHub user: %s\n", *user.Login)
	return nil
}

// CreateRepository creates a new repository in the authenticated user's account
func (c *Client) CreateRepository(name string, description string, private bool) (*github.Repository, error) {
	ctx := context.Background()
	repo := &github.Repository{
		Name:        github.String(name),
		Description: github.String(description),
		Private:     github.Bool(private),
	}
	createdRepo, _, err := c.client.Repositories.Create(ctx, "", repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %v", err)
	}
	return createdRepo, nil
}

// AddDeployKey adds a deploy key to the specified repository
func (c *Client) AddDeployKey(owner, repo, title, key string, readOnly bool) (*github.Key, error) {
    ctx := context.Background()
    deployKey := &github.Key{
        Title:    github.String(title),
        Key:      github.String(key),
        ReadOnly: github.Bool(readOnly),
    }
    createdKey, _, err := c.client.Repositories.CreateKey(ctx, owner, repo, deployKey)
    if err != nil {
        return nil, fmt.Errorf("failed to add deploy key: %v", err)
    }
    return createdKey, nil
}