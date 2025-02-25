package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

// getRelevantPostList filters out the unwanted posts and return the postList with the relevant posts
// because model.PostList.Posts contains the searched posts AND all the posts of all the linked threads
func getRelevantPostList(postList *model.PostList) *model.PostList {
	relevantPosts := make(map[string]*model.Post, len(postList.Order))

	for _, postID := range postList.Order {
		relevantPosts[postID] = postList.Posts[postID]
	}

	postList.Posts = relevantPosts
	return postList
}

type deletePostResult struct {
	numPostsDeleted    int
	technicalErrors    int
	notPermittedErrors int
	pinnedPostErrors   int
}

func (result *deletePostResult) String() (strResponse string) {
	if result.technicalErrors > 0 {
		strResponse += fmt.Sprintf(
			"Because of a technical error, %d post%s could not be deleted.\n",
			result.technicalErrors, getPluralChar(result.technicalErrors),
		)
	}

	if result.pinnedPostErrors > 0 {
		strResponse += fmt.Sprintf(
			"%d post%s not deleted because they are pinned to the channel.\n",
			result.pinnedPostErrors, getPluralChar(result.pinnedPostErrors),
		)
	}

	if result.notPermittedErrors > 0 {
		if result.numPostsDeleted == 0 {
			strResponse += "Sorry, you are only allowed to delete your own posts\n"
		} else {
			strResponse += fmt.Sprintf(
				"%d post%s not deleted because you are not allowed to do so.\n",
				result.notPermittedErrors, getPluralChar(result.notPermittedErrors),
			)
		}
	}

	if result.numPostsDeleted > 0 {
		strResponse += fmt.Sprintf(
			"Successfully deleted %d post%s.",
			result.numPostsDeleted, getPluralChar(result.numPostsDeleted))
	}

	if strResponse == "" {
		strResponse = "There are no posts in this channel."
	}

	return strResponse
}

// deletePosts deletes all the posts in postList that matches the criteria of options
// This assumes the user has the rights to delete posts
// ! This check has to be made before!
func (p *Plugin) deletePosts(postList *model.PostList, options *deletionOptions) *deletePostResult {
	p.API.LogInfo("Batch deleting these posts", "postIds", postList.Order)
	result := new(deletePostResult)

	for _, postID := range postList.Order {
		post := postList.Posts[postID]

		if !options.permDeleteOthersPosts && post.UserId != options.userID {
			result.notPermittedErrors++
			continue // process next post
		}

		if !options.optDeletePinnedPosts && post.IsPinned {
			result.pinnedPostErrors++
			continue // process next post
		}

		if post.RootId != "" {
			// The post is in a thread: skip it if the root will be also deleted,
			// because deleting a root post automatically delete the whole thread
			if _, ok := postList.Posts[post.RootId]; ok {
				result.numPostsDeleted++
				continue // process next post
			}
		}

		if appErr := p.API.DeletePost(post.Id); appErr != nil {
			result.technicalErrors++
			p.API.LogError("Unable to delete post", "PostID", post.Id, "appErr", appErr)
			continue // process next post
		}

		// FIXME Count is not accurate if we delete posts having children, but if we do not delete the children before
		// for example when using filters.
		// We can't use post.ReplyCount as a reference because it's not populated when using
		// the API
		result.numPostsDeleted++
	}

	return result
}
