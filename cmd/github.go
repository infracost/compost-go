package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"compost/internal/comment"
)

// githubCmdHandler processes common args and flags for all GitHub commands
// and returns the comment handler for posting/retrieving GitHub comments
func githubCmdHandler(ctx context.Context, cmd *cobra.Command, args []string) (*comment.CommentHandler, error) {
	project, targetType, targetRef, err := processArgs(args)
	if err != nil {
		return nil, err
	}

	apiURL, _ := cmd.Flags().GetString("github-api-url")
	token, _ := cmd.Flags().GetString("github-token")

	extra := comment.GitHubExtra{
		APIURL: apiURL,
		Token:  token,
	}

	return cmdHandler(ctx, cmd, "github", project, targetType, targetRef, extra)
}

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Post a comment to a GitHub pull request or commit",
	Example: `
  • Update a comment on a pull request:
      $ compost github update infracost/compost-example pull-request 3 --body="my comment"

  • Update a comment on a commit:
      $ compost github update infracost/compost-example commit 2ca7182 --body="my comment"`,
}

// githubUpdateCmd represents the github update command
var githubUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a comment on a GitHub pull request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(githubCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.UpdateComment(ctx, body)
	}),
}

// githubNewCmd represents the github new command
var githubNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new comment on a GitHub pull request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(githubCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.NewComment(ctx, body)
	}),
}

// githubHideAndNewCmd represents the github hide-and-new command
var githubHideAndNewCmd = &cobra.Command{
	Use:   "hide-and-new",
	Short: "Hide existing comments and create a new comment on a GitHub pull request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(githubCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.HideAndNewComment(ctx, body)
	}),
}

// githubDeleteAndNewCmd represents the github delete-and-new command
var githubDeleteAndNewCmd = &cobra.Command{
	Use:   "delete-and-new",
	Short: "Delete existing comments and create a new comment on a GitHub pull request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(githubCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.DeleteAndNewComment(ctx, body)
	}),
}

// githubLatestCmd represents the github latest command
var githubLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Return the latest comment on a GitHub pull request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: getCommentRunE(githubCmdHandler, func(ctx context.Context, handler *comment.CommentHandler) (comment.Comment, error) {
		return handler.LatestMatchingComment(ctx)
	}),
}

func init() {
	rootCmd.AddCommand(githubCmd)

	githubCmd.PersistentFlags().String("tag", "", "Customize the embedded tag that is used for detecting comments posted by Compost")
	githubCmd.PersistentFlags().String("github-api-url", "", "GitHub API URL, defaults to https://api.github.com")
	githubCmd.PersistentFlags().String("github-token", "", "GitHub token")

	githubCmd.AddCommand(githubUpdateCmd)
	githubCmd.AddCommand(githubNewCmd)
	githubCmd.AddCommand(githubHideAndNewCmd)
	githubCmd.AddCommand(githubDeleteAndNewCmd)
	githubCmd.AddCommand(githubLatestCmd)

	// Add the body and body-file flags to any commands that post comments
	for _, cmd := range []*cobra.Command{githubUpdateCmd, githubNewCmd, githubHideAndNewCmd, githubDeleteAndNewCmd} {
		cmd.Flags().String("body", "", "Body of comment to post, mutually exclusive with body-file")
		cmd.Flags().String("body-file", "", "File containing body of comment to post, mutually exclusive with body")
	}
}
