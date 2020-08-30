package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

// model.PostList.Posts contains the searched posts AND all the posts of all the linked threads
// This method filters out the unwanted posts and return a map with the relevant ones
func getRelevantPostMap(postList *model.PostList) map[string]*model.Post {
	relevantPosts := make(map[string]*model.Post, len(postList.Order))

	for _, postID := range postList.Order {
		relevantPosts[postID] = postList.Posts[postID]
	}

	return relevantPosts
}

type deletePostOptions struct {
	claimerID         string
	channelID         string
	deletePinnedPosts bool
	deleteOthersPosts bool
}

type deletePostResult struct {
	postsDeleted       int
	technicalErrors    int
	notPermittedErrors int
	pinnedPostErrors   int
}

// Assuming the user has the rights to delete the posts
// ! This check has to be made before!
func (p *Plugin) deletePostsAndTellUser(postList *model.PostList, options *deletePostOptions) *deletePostResult {
	result := new(deletePostResult)

	for _, postID := range postList.Order {
		post, ok := postList.Posts[postID]
		if !ok {
			result.technicalErrors++
			p.API.LogError("This postID doesn't match any stored post. This shouldn't happen.", "postID", postID)
			continue
		}

		if !options.deleteOthersPosts && post.UserId != options.claimerID {
			result.notPermittedErrors++
			continue // process next post
		}

		if options.deletePinnedPosts && post.IsPinned {
			result.pinnedPostErrors++
			continue // process next post
		}

		if post.RootId != "" {
			// The post is in a thread: skip it if the root will be also deleted
			if _, ok := postList.Posts[post.RootId]; ok {
				result.postsDeleted++
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

		result.postsDeleted++
	}

	return result
}

func getResponseStringFromResults(result *deletePostResult) string {
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
		if result.postsDeleted == 0 {
			strResponse += "Sorry, you are only allowed to delete your own posts\n"
		} else {
			strResponse += fmt.Sprintf(
				"%d post%s not deleted because you are not allowed to do so\n",
				result.notPermittedErrors,
				getPluralChar(result.notPermittedErrors),
			)
		}
	}

	if result.postsDeleted > 0 {
		strResponse += fmt.Sprintf(
			"Successfully deleted %d post%s",
			result.postsDeleted,
			getPluralChar(result.postsDeleted),
		)
	}

	if strResponse == "" {
		strResponse = "There are no posts in this channel"
	}

	return strResponse
}
