// pkg/github/github.go
package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// Rest of the file remains the same...

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

// AddDeployKey adds a deploy key to the specified repository
func (c *Client) AddDeployKey(repo, title, key string, readOnly bool) error {
	ctx := context.Background()
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	_, _, err := c.client.Repositories.CreateKey(ctx, owner, repoName, &github.Key{
		Title:    github.String(title),
		Key:      github.String(key),
		ReadOnly: github.Bool(readOnly),
	})
	if err != nil {
		return fmt.Errorf("failed to add deploy key: %v", err)
	}
	return nil
}