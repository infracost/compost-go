package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"compost/internal/comment"
	"compost/internal/detect"
)

// autodetectCmdHandler processes the flags and args for the autodetect commands
// and returns the comment handler for posting/retrieving comments on the detected platform
func autodetectCmdHandler(ctx context.Context, cmd *cobra.Command, args []string) (*comment.CommentHandler, error) {
	platformVal, _ := cmd.Flags().GetString("platform")
	platform, err := processPlatform(platformVal, true)
	if err != nil {
		return nil, err
	}

	targetTypeVal, _ := cmd.Flags().GetString("target-type")
	targetType, err := processTargetType(targetTypeVal, true)
	if err != nil {
		return nil, err
	}

	detectResult, err := detect.DetectEnvironment(ctx, detect.DetectOptions{
		Platform:   platform,
		TargetType: targetType,
	})
	if err != nil {
		return nil, err
	}

	return cmdHandler(
		ctx,
		cmd,
		detectResult.Platform,
		detectResult.Project,
		detectResult.TargetType,
		detectResult.TargetRef,
		detectResult.Extra,
	)
}

// autodetectCmd represents the autodetect command
var autodetectCmd = &cobra.Command{
	Use:   "autodetect",
	Short: "Post a comment to a pull/merge request or commit",
	Example: `
  • Update the previously posted comment, or create if it doesn't exist:
      $ compost autodetect update --body="my comment"

  • Post a new comment:
      $ compost autodetect new --body="my new comment"

  • Delete the previous posted comments and post a new comment:
      $ compost autodetect delete-and-new --body="my new comment"

  • Hide the previous posted comments and post a new comment (GitHub only):
      $ compost autodetect hide-and-new --body="my new comment"`,
}

// autodetectUpdateCmd represents the autodetect update command
var autodetectUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a comment on the pull/merge request or commit",
	RunE: postCommentRunE(autodetectCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.UpdateComment(ctx, body)
	}),
}

// autodetectNewCmd represents the autodetect new command
var autodetectNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new comment on the pull/merge request or commit",
	RunE: postCommentRunE(autodetectCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.NewComment(ctx, body)
	}),
}

// autodetectHideAndNewCmd represents the autodetect hide-and-new command
var autodetectHideAndNewCmd = &cobra.Command{
	Use:   "hide-and-new",
	Short: "Hide existing comments and create a new comment on the pull/merge request or commit",
	RunE: postCommentRunE(autodetectCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.HideAndNewComment(ctx, body)
	}),
}

// autodetectDeleteAndNewCmd represents the autodetect delete-and-new command
var autodetectDeleteAndNewCmd = &cobra.Command{
	Use:   "delete-and-new",
	Short: "Delete existing comments and create a new comment on the pull/merge request or commit",
	RunE: postCommentRunE(autodetectCmdHandler, func(ctx context.Context, handler *comment.CommentHandler, body string) error {
		return handler.DeleteAndNewComment(ctx, body)
	}),
}

// autodetectLatestCmd represents the autodetect latest command
var autodetectLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Return the latest comment on the pull/merge request or commit",
	RunE: getCommentRunE(autodetectCmdHandler, func(ctx context.Context, handler *comment.CommentHandler) (comment.Comment, error) {
		return handler.LatestMatchingComment(ctx)
	}),
}

func init() {
	rootCmd.AddCommand(autodetectCmd)

	autodetectCmd.PersistentFlags().String("tag", "", "Customize the embedded tag that is used for detecting comments posted by Compost")
	autodetectCmd.PersistentFlags().String("platform", "", "Limit the auto-detection to a specific platform: github, gitlab")
	autodetectCmd.PersistentFlags().String("target-type", "", "Limit the auto-detection to pull/merge requests or commits: pull-request (pr), merge-request (mr), commit")

	autodetectCmd.AddCommand(autodetectUpdateCmd)
	autodetectCmd.AddCommand(autodetectNewCmd)
	autodetectCmd.AddCommand(autodetectHideAndNewCmd)
	autodetectCmd.AddCommand(autodetectDeleteAndNewCmd)
	autodetectCmd.AddCommand(autodetectLatestCmd)

	// Add the body and body-file flags to any commands that post comments
	for _, cmd := range []*cobra.Command{autodetectUpdateCmd, autodetectNewCmd, autodetectHideAndNewCmd, autodetectDeleteAndNewCmd} {
		cmd.Flags().String("body", "", "Body of comment to post, mutually exclusive with body-file")
		cmd.Flags().String("body-file", "", "File containing body of comment to post, mutually exclusive with body")
	}
}
