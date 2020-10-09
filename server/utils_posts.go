package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

// model.PostList.Posts contains the searched posts AND all the posts of all the linked threads
// This method filters out the unwanted posts and return the postList with the relevant posts
func getRelevantPostList(postList *model.PostList) *model.PostList {
	relevantPosts := make(map[string]*model.Post, len(postList.Order))

	for _, postID := range postList.Order {
		relevantPosts[postID] = postList.Posts[postID]
	}

	postList.Posts = relevantPosts
	return postList
}

type delOptions struct {
	userID                string
	channelID             string
	numPost               int
	optDeletePinnedPosts  bool
	optNoConfirmDialog    bool
	permDeleteOthersPosts bool
}

type delResults struct {
	numPostsDeleted    int
	technicalErrors    int
	notPermittedErrors int
	pinnedPostErrors   int
}

// Assuming the user has the rights to delete the posts
// ! This check has to be made before!
func (p *Plugin) deletePosts(postList *model.PostList, options *delOptions) *delResults {
	p.API.LogDebug("Deleting these posts", "postIds", postList.Order)
	result := new(delResults)

	for _, postID := range postList.Order {
		post, ok := postList.Posts[postID]
		if !ok {
			result.technicalErrors++
			p.API.LogError("This postID doesn't match any stored post. This shouldn't happen.", "postID", postID)
			continue
		}

		if !options.permDeleteOthersPosts && post.UserId != options.userID {
			result.notPermittedErrors++
			continue // process next post
		}

		if options.optDeletePinnedPosts && post.IsPinned {
			result.pinnedPostErrors++
			continue // process next post
		}

		if post.RootId != "" {
			// The post is in a thread: skip it if the root will be also deleted
			if _, ok := postList.Posts[post.RootId]; ok {
				result.numPostsDeleted++
				continue
			}
		}

		if appErr := p.API.DeletePost(post.Id); appErr != nil {
			result.technicalErrors++
			p.API.LogError(
				"Unable to delete post",
				"PostID", post.Id,
				"err", appErr.Error(),
				"ErrorId", appErr.Id,
				"RequestId", appErr.RequestId,
				"DetailedError", appErr.DetailedError,
			)
			continue
		}

		result.numPostsDeleted++
	}

	return result
}

func getResponseStringFromResults(result *delResults) string {
	strResponse := ""

	if result.technicalErrors > 0 {
		strResponse += fmt.Sprintf(
			"Because of a technical error, %d post%s could not be deleted\n",
			result.technicalErrors,
			getPluralChar(result.technicalErrors),
		)
	}

	if result.pinnedPostErrors > 0 {
		strResponse += fmt.Sprintf(
			"%d post%s not deleted because pinned to channel\n",
			result.pinnedPostErrors,
			getPluralChar(result.pinnedPostErrors),
		)
	}

	if result.notPermittedErrors > 0 {
		if result.numPostsDeleted == 0 {
			strResponse += "Sorry, you are only allowed to delete your own posts\n"
		} else {
			strResponse += fmt.Sprintf(
				"%d post%s not deleted because you are not allowed to do so\n",
				result.notPermittedErrors,
				getPluralChar(result.notPermittedErrors),
			)
		}
	}

	if result.numPostsDeleted > 0 {
		strResponse += fmt.Sprintf(
			"Successfully deleted %d post%s",
			result.numPostsDeleted,
			getPluralChar(result.numPostsDeleted),
		)
	}

	if strResponse == "" {
		strResponse = "No post matches your filters: this channel looks clean!"
	}

	return strResponse
}
