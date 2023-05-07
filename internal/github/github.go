/*
Copyright © 2023 Kong, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// Options contain the parameters to create a new GitHub client instance.
type Options struct {
	// Token represents the GITHUB_TOKEN for the client
	Token string
	// Logger represents the base logger to use for the GitHub package
	Logger *zap.Logger
}

// Client represents a GitHub client instance.
type Client struct {
	// client represents the connection for GitHub's API
	client *github.Client
	// logger represents the logger to use for the GitHub package
	logger *zap.Logger
}

// NewClient will validate options and instantiate a new GitHub client instance.
func NewClient(opts Options) (*Client, error) {
	// Validate required options
	if len(strings.TrimSpace(opts.Token)) == 0 {
		return nil, errors.New("github token is not set")
	}
	if opts.Logger == nil {
		return nil, errors.New("logger is not set")
	}

	// Create the HTTP client using the GITHUB_TOKEN provided
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Token})
	httpClient := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(httpClient.Transport)
	if err != nil {
		return nil, fmt.Errorf("unable to create GitHub rate limiter: %w", err)
	}

	return &Client{
		client: github.NewClient(rateLimiter),
		logger: opts.Logger.With(zap.String("component", "github")),
	}, nil
}

// PRDescription will get the description of a pull request for a given
// organization and repository.
func (c *Client) PRDescription(organization string, repository string, pullRequest int) (string, error) {
	// Get the pull request for the given parameters
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pr, _, err := c.client.PullRequests.Get(ctx, organization, repository, pullRequest)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve pull request: %w", err)
	}

	return *pr.Body, nil
}
