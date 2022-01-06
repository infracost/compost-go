package detect

import (
	"compost/internal/comment"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

// GitHubActionsDetector detects a GitHub Actions environment.
type GitHubActionsDetector struct{}

// DisplayName is the display name to use in any logs or output for this detector.
func (d *GitHubActionsDetector) DisplayName() string {
	return "GitHub Actions"
}

// Detect checks the environment variables to determine if it is running in
// a GitHub Actions environment. If it is it returns a DetectResult, otherwise
// it throws a DetectError.
//
// If the action is running in the context of a pull request it returns a
// target type of pull-request and the the pull request number as the target ref.
// Otherwise it returns a target type of commit and the commit SHA.
func (d *GitHubActionsDetector) Detect(ctx context.Context, opts DetectOptions) (DetectResult, error) {
	err := checkEnvVarValue(ctx, "GITHUB_ACTIONS", "true", false)
	if err != nil {
		return DetectResult{}, &DetectError{err}
	}

	token, err := checkEnvVarExists(ctx, "GITHUB_TOKEN", true)
	if err != nil {
		return DetectResult{}, &DetectError{err}
	}

	project, err := checkEnvVarExists(ctx, "GITHUB_REPOSITORY", false)
	if err != nil {
		return DetectResult{}, &DetectError{err}
	}

	apiURL := os.Getenv("GITHUB_API_URL")

	eventPath := os.Getenv("GITHUB_EVENT_PATH")

	var event struct {
		PullRequest struct {
			Number int
			Head   struct {
				SHA string
			}
		} `json:"pull_request"`
	}

	if eventPath != "" {
		content, err := os.ReadFile(eventPath)
		if err != nil {
			return DetectResult{}, &DetectError{err}
		}
		err = json.Unmarshal(content, &event)
		if err != nil {
			return DetectResult{}, &DetectError{err}
		}
	}

	var targetType string
	var targetRef string

	if opts.TargetType == "" || opts.TargetType == "pull-request" {
		targetRef = strconv.Itoa(event.PullRequest.Number)
		if targetRef != "" {
			targetType = "pull-request"
		}
	}

	if targetRef == "" && opts.TargetType == "" || opts.TargetType == "commit" {
		targetType = "commit"
		targetRef = event.PullRequest.Head.SHA
		if targetRef == "" {
			targetRef, err = checkEnvVarExists(ctx, "GITHUB_SHA", false)
			if err != nil {
				return DetectResult{}, &DetectError{err}
			}
		}
	}

	if targetRef == "" {
		return DetectResult{}, &DetectError{errors.New("Could not determine target ref")}
	}

	return DetectResult{
		Platform:   "github",
		Project:    project,
		TargetType: targetType,
		TargetRef:  targetRef,
		Extra: comment.GitHubExtra{
			APIURL: apiURL,
			Token:  token,
		},
	}, nil
}

func init() {
	// Here we register the detectors against the platforms they detect
	registerDetector([]string{"github"}, &GitHubActionsDetector{})
}
