package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

type deletePostOptions struct {
	claimerID         string
	channelID         string
	deletePinnedPosts bool
	deleteOthersPosts bool
	// TODO: add others options
}

// Assuming the user has the rights to delete the posts
// !This check has to be made before!
func (p *Plugin) deletePostsAndTellUser(postsToDelete map[string]*model.Post, options *deletePostOptions, ephemeralPostBeginning *model.Post) {
	const (
		DeletePosts int = iota
		TechnicalError
		NotPermittedError
		PinnedPostError
	)
	result := make(map[int]int64, 4)

	for _, post := range postsToDelete {
		if !options.deleteOthersPosts && post.UserId != options.claimerID {
			result[NotPermittedError]++
			continue // process next post
		}

		if options.deletePinnedPosts && post.IsPinned {
			result[PinnedPostError]++
			continue // process next post
		}

		if post.RootId != "" {
			// The post is in a thread: skip it if the root will be also deleted
			if _, ok := postsToDelete[post.RootId]; ok {
				result[DeletePosts]++
				continue
			}
		}

		if appErr := p.API.DeletePost(post.Id); appErr != nil {
			result[TechnicalError]++
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

		result[DeletePosts]++
	}

	strResponse := ""

	if result[TechnicalError] > 0 {
		strResponse += fmt.Sprintf(
			"Because of a technical error, %d post%s could not be deleted\n",
			result[TechnicalError],
			getPluralChar(result[TechnicalError]),
		)
	}

	if result[PinnedPostError] > 0 {
		strResponse += fmt.Sprintf(
			"%d post%s not deleted because pinned to channel\n",
			result[PinnedPostError],
			getPluralChar(result[PinnedPostError]),
		)
	}

	if result[NotPermittedError] > 0 {
		if result[DeletePosts] == 0 {
			strResponse += "Sorry, you are only allowed to delete your own posts\n"
		} else {
			strResponse += fmt.Sprintf(
				"%d post%s not deleted because you are not allowed to do so\n",
				result[NotPermittedError],
				getPluralChar(result[NotPermittedError]),
			)
		}
	}

	if result[DeletePosts] > 0 {
		strResponse += fmt.Sprintf(
			"Successfully deleted %d post%s",
			result[DeletePosts],
			getPluralChar(result[DeletePosts]),
		)
	}

	if strResponse == "" {
		strResponse = "There are no posts in this channel"
	}

	p.API.DeleteEphemeralPost(options.claimerID, ephemeralPostBeginning.Id)
	p.sendEphemeralPost(options.claimerID, options.channelID, strResponse)
}
