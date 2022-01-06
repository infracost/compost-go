package cmd

import (
	"compost/internal/comment"
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type commentHandlerFactory func(ctx context.Context, cmd *cobra.Command, args []string) (*comment.CommentHandler, error)
type postCommentFunc func(ctx context.Context, handler *comment.CommentHandler, body string) error
type getCommentFunc func(ctx context.Context, handler *comment.CommentHandler) (comment.Comment, error)

// processArgs process the common args for all commands that post and get comments for a platform.
func processArgs(args []string) (string, string, string, error) {
	if len(args) != 3 {
		return "", "", "", fmt.Errorf("Expected 3 arguments (project, target-type, target-ref), got %d", len(args))
	}

	targetType, err := processTargetType(args[1], false)
	if err != nil {
		return "", "", "", err
	}

	return args[0], targetType, args[2], nil
}

// processTargetType maps the target type to the target type used by the comment handler.
// It returns an error if the target type is not supported.
func processTargetType(s string, allowEmpty bool) (string, error) {
	if allowEmpty && s == "" {
		return "", nil
	}

	v, ok := map[string]string{
		"pr":            "pull-request",
		"pull-request":  "pull-request",
		"mr":            "pull-request",
		"merge-request": "pull-request",
		"commit":        "commit",
	}[s]

	if !ok {
		return "", fmt.Errorf("Invalid target type '%s', valid options are 'pull-request' ('pr'), 'merge-request' ('mr'), 'commit'", s)
	}

	return v, nil
}

// processPlatform maps the platform to the target type used by the comment handler.
// It returns an error if the platform is not supported.
func processPlatform(s string, allowEmpty bool) (string, error) {
	if allowEmpty && s == "" {
		return "", nil
	}

	v, ok := map[string]string{
		"github": "github",
		"gitlab": "gitlab",
		"":       "",
	}[s]

	if !ok {
		return "", fmt.Errorf("Invalid platform '%s', valid options are 'github', 'gitlab'", s)
	}

	return v, nil
}

// processBodyFlags processes the body and body-file flags and returns the body.
// It returns an error if neither or both are set.
// If body-file is set it reads the contents of the body file.
func processBodyFlags(cmd *cobra.Command) (string, error) {
	bodySet := cmd.Flags().Changed("body")
	bodyFileSet := cmd.Flags().Changed("body-file")

	if !bodySet && !bodyFileSet {
		return "", fmt.Errorf("--body or --body-file must be set")
	}

	if bodySet && bodyFileSet {
		return "", fmt.Errorf("--body and --body-file cannot be set at the same time")
	}

	if bodySet {
		body, _ := cmd.Flags().GetString("body")
		return body, nil
	}

	bodyFile, _ := cmd.Flags().GetString("body-file")
	b, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", errors.Wrap(err, "Failed to read body file")
	}

	return string(b), nil
}

// cmdHandler processes common args for all commands
// and returns the comment handler for posting/retrieving comments on the given platform.
func cmdHandler(ctx context.Context, cmd *cobra.Command, platform string, project string, targetType string, targetRef string, extra interface{}) (*comment.CommentHandler, error) {
	tag, _ := cmd.Flags().GetString("tag")

	platformHandlerFactory, err := comment.NewPlatformHandlerFactory(ctx, platform, targetType)
	if err != nil {
		return nil, err
	}

	platformHandler, err := platformHandlerFactory(ctx, project, targetRef, extra)
	if err != nil {
		return nil, err
	}

	return &comment.CommentHandler{
		PlatformHandler: platformHandler,
		Tag:             tag,
	}, nil
}

// postCommentRunE contains the common logic for any command that posts comments.
// It sets up the logger, creates the comment handler, processes the args and flags
// and calls the handlerFunc to post the comment.
func postCommentRunE(handlerFactory commentHandlerFactory, handlerFunc postCommentFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		handler, err := handlerFactory(ctx, cmd, args)
		if err != nil {
			return err
		}

		body, err := processBodyFlags(cmd)
		if err != nil {
			return err
		}

		return handlerFunc(ctx, handler, body)
	}
}

// getCommentRunE contains the common logic for any command that gets comments.
// It sets up the logger, creates the comment handler, processes the args and flags,
// calls the handlerFunc to retrieve the comment and outputs the comment to stdout.
func getCommentRunE(handlerFactory commentHandlerFactory, handlerFunc getCommentFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		handler, err := handlerFactory(ctx, cmd, args)
		if err != nil {
			return err
		}

		comment, err := handlerFunc(ctx, handler)
		if err != nil {
			return err
		}

		if comment.Body() != "" {
			cmd.Println(comment.Body())
		}

		return nil
	}
}
