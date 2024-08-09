package gitlab // assume this is the package name, adjust if different

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

// GitLabClient represents a GitLab client
type GitLabClient struct {
	client *gitlab.Client
}

// NewGitLabClient creates a new GitLab client with the provided token
func NewGitLabClient(token string) (*GitLabClient, error) {
	if token == "" {
		return nil, fmt.Errorf("GitLab token is required")
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %v", err)
	}

	return &GitLabClient{client: client}, nil
}

// Authenticate verifies the GitLab token by fetching the authenticated user
func (c *GitLabClient) Authenticate() error {
	user, _, err := c.client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}
	fmt.Printf("Authenticated as GitLab user: %s\n", user.Username)
	return nil
}

// AddDeployKey adds a deploy key to the specified repository
func (c *GitLabClient) AddDeployKey(projectID int, title, key string, readOnly bool) error {
	_, _, err := c.client.DeployKeys.AddDeployKey(projectID, &gitlab.AddDeployKeyOptions{
		Title:    gitlab.String(title),
		Key:      gitlab.String(key),
		CanPush:  gitlab.Bool(!readOnly),
		ReadOnly: gitlab.Bool(readOnly),
	})
	if err != nil {
		return fmt.Errorf("failed to add deploy key: %v", err)
	}
	return nil
}