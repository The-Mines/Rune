// pkg/gitlab/gitlab.go
package gitlab

import (
	"fmt"
	"strconv"

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
func (c *GitLabClient) AddDeployKey(projectID, title, key string, readOnly bool) error {
	// Try to convert projectID to int, if it fails, assume it's a project path
	id, err := strconv.Atoi(projectID)
	if err != nil {
		// If projectID is not a number, treat it as a project path
		project, _, err := c.client.Projects.GetProject(projectID, &gitlab.GetProjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to get project: %v", err)
		}
		id = project.ID
	}

	_, _, err = c.client.DeployKeys.AddDeployKey(id, &gitlab.AddDeployKeyOptions{
		Title:   gitlab.String(title),
		Key:     gitlab.String(key),
		CanPush: gitlab.Bool(!readOnly),
	})
	if err != nil {
		return fmt.Errorf("failed to add deploy key: %v", err)
	}
	return nil
}