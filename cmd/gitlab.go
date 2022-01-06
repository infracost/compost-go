package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"compost/internal/comment"
)

// gitlabCmdHandler processes common args and flags for all GitLab commands
// and returns the comment handler for posting/retrieving GitLab comments
func gitlabCmdHandler(ctx context.Context, cmd *cobra.Command, args []string) (*comment.CommentHandler, error) {
	project, targetType, targetRef, err := processArgs(args)
	if err != nil {
		return nil, err
	}

	serverURL, _ := cmd.Flags().GetString("gitlab-server-url")
	token, _ := cmd.Flags().GetString("gitlab-token")

	extra := comment.GitLabExtra{
		ServerURL: serverURL,
		Token:     token,
	}

	return cmdHandler(ctx, cmd, "gitlab", project, targetType, targetRef, extra)
}

// gitlabCmd represents the gitlab command
var gitlabCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "Post a comment to a GitLab merge request or commit",
	Example: `
  • Update a comment on a merge request:
      $ compost gitlab update infracost/compost-example merge-request 3 --body="my comment"

  • Update a comment on a commit:
      $ compost gitlab update infracost/compost-example commit 2ca7182 --body="my comment"`,
}

// gitlabUpdateCmd represents the gitlab update command
var gitlabUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a comment on a GitLab merge request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(gitlabCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.UpdateComment(ctx, body)
	}),
}

// gitlabNewCmd represents the gitlab new command
var gitlabNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new comment on a GitLab merge request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(gitlabCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.NewComment(ctx, body)
	}),
}

// gitlabDeleteAndNewCmd represents the gitlab delete-and-new command
var gitlabDeleteAndNewCmd = &cobra.Command{
	Use:   "delete-and-new",
	Short: "Delete existing comments and create a new comment on a GitLab merge request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: postCommentRunE(gitlabCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.DeleteAndNewComment(ctx, body)
	}),
}

// gitlabLatestCmd represents the gitlab latest command
var gitlabLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Return the latest comment on a GitLab merge request or commit",
	Args:  cobra.ExactValidArgs(3),
	RunE: getCommentRunE(gitlabCmdHandler, func(ctx context.Context, handler *comment.CommentHandler) (comment.Comment, error) {
		return handler.LatestMatchingComment(ctx)
	}),
}

func init() {
	rootCmd.AddCommand(gitlabCmd)

	gitlabCmd.PersistentFlags().String("tag", "", "Customize the embedded tag that is used for detecting comments posted by Compost")
	gitlabCmd.PersistentFlags().String("gitlab-server-url", "", "GitLab server URL, defaults to https://gitlab.com")
	gitlabCmd.PersistentFlags().String("gitlab-token", "", "GitLab token")

	gitlabCmd.AddCommand(gitlabUpdateCmd)
	gitlabCmd.AddCommand(gitlabNewCmd)
	gitlabCmd.AddCommand(gitlabDeleteAndNewCmd)
	gitlabCmd.AddCommand(gitlabLatestCmd)

	// Add the body and body-file flags to any commands that post comments
	for _, cmd := range []*cobra.Command{gitlabUpdateCmd, gitlabNewCmd, gitlabDeleteAndNewCmd} {
		cmd.Flags().String("body", "", "Body of comment to post, mutually exclusive with body-file")
		cmd.Flags().String("body-file", "", "File containing body of comment to post, mutually exclusive with body")
	}
}
