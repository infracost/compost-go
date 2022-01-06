package detect

import (
	"compost/internal/comment"
	"context"
	"errors"
	"os"
)

// GitLabCIDetector detects a GitLab CI environment.
type GitLabCIDetector struct{}

// DisplayName is the display name to use in any logs or output for this detector.
func (d *GitLabCIDetector) DisplayName() string {
	return "GitLab CI"
}

// Detect checks the environment variables to determine if it is running in
// a GitLab CI environment. If it is it returns a DetectResult, otherwise
// it throws a DetectError.
//
// If the action is running in the context of a merge request it returns a
// target type of pull-request and the the merge request number as the target ref.
// Otherwise it returns a target type of commit and the commit SHA.
func (d *GitLabCIDetector) Detect(ctx context.Context, opts DetectOptions) (DetectResult, error) {
	err := checkEnvVarValue(ctx, "GITLAB_CI", "true", false)
	if err != nil {
		return DetectResult{}, &DetectError{err}
	}

	token, err := checkEnvVarExists(ctx, "GITLAB_TOKEN", true)
	if err != nil {
		return DetectResult{}, &DetectError{err}
	}

	project, err := checkEnvVarExists(ctx, "CI_PROJECT_PATH", false)
	if err != nil {
		return DetectResult{}, &DetectError{err}
	}

	serverURL := os.Getenv("CI_SERVER_URL")

	var targetType string
	var targetRef string

	if opts.TargetType == "" || opts.TargetType == "pull-request" {
		targetRef = os.Getenv("CI_MERGE_REQUEST_IID")
		if targetRef != "" {
			targetType = "pull-request"
		}
	}

	if targetRef == "" && opts.TargetType == "" || opts.TargetType == "commit" {
		targetType = "commit"
		targetRef, err = checkEnvVarExists(ctx, "CI_COMMIT_SHA", false)
		if err != nil {
			return DetectResult{}, &DetectError{err}
		}
	}

	if targetRef == "" {
		return DetectResult{}, &DetectError{errors.New("Could not determine target ref")}
	}

	return DetectResult{
		Platform:   "gitlab",
		Project:    project,
		TargetType: targetType,
		TargetRef:  targetRef,
		Extra: comment.GitLabExtra{
			ServerURL: serverURL,
			Token:     token,
		},
	}, nil
}

func init() {
	// Here we register the detectors against the platforms they detect
	registerDetector([]string{"gitlab"}, &GitLabCIDetector{})
}
