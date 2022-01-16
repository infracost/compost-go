package comment

import (
	"context"
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

var defaultTag = "compost-comment"

// Comment is an interface that represents a comment on any platform. It wraps
// the platform specific comment structures and is used to abstract the
// logic for finding, creating, updating, and deleting the comments.
type Comment interface {
	// Body returns the body of the comment.
	Body() string

	// Ref returns the reference of the comment, this can be a URL to the HTML page of the comment.
	Ref() string

	// Less compares the comment to another comment and returns true if this
	// comment should be sorted before the other comment.
	Less(c Comment) bool

	// IsHidden returns true if the comment is hidden or minimized.
	IsHidden() bool
}

// PlatformHandler is an interface that represents a platform specific handler.
// It is used to call the platform-specific APIs for finding, creating, updating
// and deleting comments.
type PlatformHandler interface {
	// CallFindMatchingComments calls the platform-specific API to find
	// comments that match the given tag, which has been embedded at the beginning
	// of the comment.
	CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error)

	// CallCreateComment calls the platform-specific API to create a new comment.
	CallCreateComment(ctx context.Context, body string) (Comment, error)

	// CallUpdateComment calls the platform-specific API to update the body of a comment.
	CallUpdateComment(ctx context.Context, comment Comment, body string) error

	// CallDeleteComment calls the platform-specific API to delete the comment.
	CallDeleteComment(ctx context.Context, comment Comment) error

	// CallHideComment calls the platform-specific API to minimize the comment.
	// This functionality is not supported by all platforms, in which case this
	// will throw a NotImplemented error.
	CallHideComment(ctx context.Context, comment Comment) error
}

// PlatformHandlerFactory is a function that creates a new PlatformHandler.
// It requires:
//  • project: either the name or URL of the repository depending on the platform
//  • targerRef: either the pull/merge request number or a commit SHA
//  • extra: any extra data that can be used to create the PlatformHandler, e.g. a token or API URl
type PlatformHandlerFactory func(
	ctx context.Context,
	project string,
	targetRef string,
	extra interface{}) (PlatformHandler, error)

// platformHandlerRegistryItem represents an item in the platform handler registry.
// It maps the platform handlers factory function to the platform and target type
// it supports.
type platformHandlerRegistryItem struct {
	platform   string
	targetType string
	factory    PlatformHandlerFactory
}

// platformHandlerRegistry contains the list of all platform handlers.
// PlatformHandlers are registered using the registerPlatformHandler function in the init()
// function of the files containing them.
var platformHandlerRegister = []platformHandlerRegistryItem{}

// registerPlatformHandler registers a new platform handler in the platform handler registry,
// mapping it to the platform and target type it supports
func registerPlatformHandler(platform string, targetType string, factory PlatformHandlerFactory) {
	platformHandlerRegister = append(platformHandlerRegister, platformHandlerRegistryItem{
		platform:   platform,
		targetType: targetType,
		factory:    factory,
	})
}

// NewPlatformHandlerFactory returns the PlatformHandlerFactory for the given platform and targetType.
func NewPlatformHandlerFactory(ctx context.Context, platform string, targetType string) (PlatformHandlerFactory, error) {
	for _, item := range platformHandlerRegister {
		if item.platform == platform && item.targetType == targetType {
			return item.factory, nil
		}
	}

	return nil, fmt.Errorf("%s (%s) is not supported", platform, targetType)
}

// CommentHandler contains the logic for finding, creating, updating and deleting comments
// on any platform. It uses a PlatformHandler to call the platform-specific APIs.
type CommentHandler struct {
	PlatformHandler PlatformHandler
	Tag             string
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(ctx context.Context, platformHandler PlatformHandler, tag string) (*CommentHandler, error) {
	if tag == "" {
		tag = defaultTag
	}

	return &CommentHandler{
		PlatformHandler: platformHandler,
		Tag:             tag,
	}, nil
}

// matchingComments returns all comments that match the tag.
func (h *CommentHandler) matchingComments(ctx context.Context) ([]Comment, error) {
	log.Ctx(ctx).Info().Msgf("Finding matching comments for tag %s", h.Tag)

	matchingComments, err := h.PlatformHandler.CallFindMatchingComments(ctx, h.Tag)
	if err != nil {
		return nil, err
	}

	if len(matchingComments) == 1 {
		log.Ctx(ctx).Info().Msg("Found 1 matching comment")
	} else {
		log.Ctx(ctx).Info().Msgf("Found %d matching comments", len(matchingComments))
	}

	return matchingComments, nil
}

// LatestMatchingComment returns the latest matching comment.
func (h *CommentHandler) LatestMatchingComment(ctx context.Context) (Comment, error) {
	matchingComments, err := h.matchingComments(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(matchingComments, func(i, j int) bool {
		return matchingComments[i].Less(matchingComments[j])
	})

	if len(matchingComments) == 0 {
		return nil, nil
	}

	return matchingComments[0], nil
}

// UpdateComment updates the comment with the given body.
func (h *CommentHandler) UpdateComment(ctx context.Context, body string) error {
	bodyWithTag := addMarkdownTag(body, h.Tag)

	latestMatchingComment, err := h.LatestMatchingComment(ctx)
	if err != nil {
		return err
	}

	if latestMatchingComment != nil {
		if latestMatchingComment.Body() == bodyWithTag {
			log.Ctx(ctx).Info().Msgf("Not updating comment since the latest one matches exactly: %s", color.HiBlueString(latestMatchingComment.Ref()))
			return nil
		}

		log.Ctx(ctx).Info().Msgf("Updating comment %s", color.HiBlueString(latestMatchingComment.Ref()))

		err := h.PlatformHandler.CallUpdateComment(ctx, latestMatchingComment, bodyWithTag)
		if err != nil {
			return err
		}
	} else {
		log.Ctx(ctx).Info().Msg("Creating new comment")

		comment, err := h.PlatformHandler.CallCreateComment(ctx, bodyWithTag)
		if err != nil {
			return err
		}

		log.Ctx(ctx).Info().Msgf("Created new comment %s", color.HiBlueString(comment.Ref()))
	}

	return nil
}

// NewComment creates a new comment with the given body.
func (h *CommentHandler) NewComment(ctx context.Context, body string) error {
	bodyWithTag := body

	log.Ctx(ctx).Info().Msg("Creating new comment")

	comment, err := h.PlatformHandler.CallCreateComment(ctx, bodyWithTag)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Msgf("Created new comment: %s", color.HiBlueString(comment.Ref()))

	return err
}

// HideAndNewComment hides/minimizes all existing matching comment and creates a new one with the given body.
func (h *CommentHandler) HideAndNewComment(ctx context.Context, body string) error {
	matchingComments, err := h.matchingComments(ctx)
	if err != nil {
		return err
	}

	err = h.hideComments(ctx, matchingComments)
	if err != nil {
		return err
	}

	return h.NewComment(ctx, body)
}

// hideComments hides/minimizes all the given comments.
func (h *CommentHandler) hideComments(ctx context.Context, comments []Comment) error {
	visibleComments := []Comment{}

	for _, comment := range comments {
		if !comment.IsHidden() {
			visibleComments = append(visibleComments, comment)
		}
	}

	hiddenCommentCount := len(comments) - len(visibleComments)

	if hiddenCommentCount == 1 {
		log.Ctx(ctx).Info().Msg("1 comments is already hidden")
	} else if hiddenCommentCount > 0 {
		log.Ctx(ctx).Info().Msgf("%d comments are already hidden", hiddenCommentCount)
	}

	if len(visibleComments) == 1 {
		log.Ctx(ctx).Info().Msg("Hiding 1 comment")
	} else {
		log.Ctx(ctx).Info().Msgf("Hiding %d comments", len(visibleComments))
	}

	for _, comment := range visibleComments {
		log.Ctx(ctx).Info().Msgf("Hiding comment %s", color.HiBlueString(comment.Ref()))
		err := h.PlatformHandler.CallHideComment(ctx, comment)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteAndNewComment deletes all existing matching comment and creates a new one with the given body.
func (h *CommentHandler) DeleteAndNewComment(ctx context.Context, body string) error {
	matchingComments, err := h.matchingComments(ctx)
	if err != nil {
		return err
	}

	err = h.deleteComments(ctx, matchingComments)
	if err != nil {
		return err
	}

	return h.NewComment(ctx, body)
}

// deleteComments hides/minimizes all the given comments.
func (h *CommentHandler) deleteComments(ctx context.Context, comments []Comment) error {
	if len(comments) == 1 {
		log.Ctx(ctx).Info().Msg("Deleting 1 comment")
	} else {
		log.Ctx(ctx).Info().Msgf("Deleting %d comments", len(comments))
	}

	for _, comment := range comments {
		log.Ctx(ctx).Info().Msgf("Deleting comment %s", color.HiBlueString(comment.Ref()))
		err := h.PlatformHandler.CallDeleteComment(ctx, comment)
		if err != nil {
			return err
		}
	}

	return nil
}
