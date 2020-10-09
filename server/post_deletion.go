package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
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

// delOptions contains the information used for deletion of posts
type delOptions struct {
	userID                string
	triggerID             string
	channelID             string
	numPost               int
	optDeletePinnedPosts  bool
	optNoConfirmDialog    bool
	permDeleteOthersPosts bool
}

// delResults contains statistics about the deleted (or not) posts
type delResults struct {
	numPostsDeleted    int
	technicalErrors    int
	notPermittedErrors int
	pinnedPostErrors   int
}

// batchDeletePosts delete all the posts in postList that matches the criteria of options
// This assumes the user has the rights to delete posts
// ! This check has to be made before!
func (p *Plugin) batchDeletePosts(postList *model.PostList, options *delOptions) *delResults {
	p.API.LogInfo("Batch deleting these posts", "postIds", postList.Order)
	result := new(delResults)

	for _, postID := range postList.Order {
		post, ok := postList.Posts[postID]
		if !ok {
			result.technicalErrors++
			p.API.LogError("This postID doesn't match any stored post. This shouldn't happen.", "postID", postID, "storedPost", postList.Posts)
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
			// The post is in a thread: skip it if the root will be also deleted,
			// because deleting a root post automatically delete the whole thread
			if _, ok := postList.Posts[post.RootId]; ok {
				result.numPostsDeleted++
				continue
			}
		}

		if appErr := p.API.DeletePost(post.Id); appErr != nil {
			result.technicalErrors++
			p.API.LogError("Unable to delete post", "PostID", post.Id, "appErr", appErr.ToJson())
			continue
		}

		result.numPostsDeleted++
	}

	return result
}

// String turns the delResults into a formatted string
func (delR *delResults) String() string {
	str := ""

	if delR.technicalErrors > 0 {
		str += fmt.Sprintf(
			"Because of a technical error, %d post%s could not be deleted\n",
			delR.technicalErrors,
			getPluralChar(delR.technicalErrors),
		)
	}

	if delR.pinnedPostErrors > 0 {
		str += fmt.Sprintf(
			"%d post%s not deleted because pinned to channel\n",
			delR.pinnedPostErrors,
			getPluralChar(delR.pinnedPostErrors),
		)
	}

	if delR.notPermittedErrors > 0 {
		if delR.numPostsDeleted == 0 {
			str += "Sorry, you are only allowed to delete your own posts\n"
		} else {
			str += fmt.Sprintf(
				"%d post%s not deleted because you are not allowed to do so\n",
				delR.notPermittedErrors,
				getPluralChar(delR.notPermittedErrors),
			)
		}
	}

	if delR.numPostsDeleted > 0 {
		str += fmt.Sprintf(
			"Successfully deleted %d post%s",
			delR.numPostsDeleted,
			getPluralChar(delR.numPostsDeleted),
		)
	}

	if str == "" {
		str = "No post matches your filters: this channel looks clean!"
	}

	return str
}
