package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) deleteLastPostsInChannel(numPostToDelete int, channelID string, userID string, deletePinnedPosts bool) *model.AppError {
	postList, err := p.API.GetPostsForChannel(channelID, 0, numPostToDelete)
	if err != nil {
		p.API.LogError(
			"Unable to retrieve posts",
			"err", err.Error(),
		)
		return err
	}

	hasPermissionToDeletePost := canDeletePost(p, userID, channelID)
	hasPermissionToDeleteOthersPosts := canDeleteOthersPosts(p, userID, channelID)

	if !hasPermissionToDeletePost && !hasPermissionToDeleteOthersPosts {
		p.sendEphemeralPost(&model.CommandArgs{
			ChannelId: channelID,
			UserId:    userID,
		}, "Sorry, you are not permitted to delete posts")
		return nil
	}

	isError := false
	isErrorNotAdmin := false
	isErrorPinnedPost := false
	numDeletedPost := 0
	for _, postID := range postList.Order {
		post, err := p.API.GetPost(postID)
		if err != nil {
			isError = true
			p.API.LogError(
				"Unable to get post "+postID+" informations",
				"err", err.Error(),
			)
			continue // process next post
		}

		if !hasPermissionToDeleteOthersPosts {
			if post.UserId != userID {
				isErrorNotAdmin = true
				continue // process next post
			}
		}

		if post.IsPinned && !deletePinnedPosts {
			isErrorPinnedPost = true
			continue // process next post
		}

		if err := p.API.DeletePost(postID); err != nil {
			isError = true
			p.API.LogError(
				"Unable to delete post",
				"post id", postID,
				"err", err.Error(),
			)
		} else {
			numDeletedPost++
		}
	}

	strResponse := ""

	if isError {
		strResponse += "An error has occurred, some post could not be deleted\n"
	}

	if isErrorNotAdmin {
		if numDeletedPost == 0 {
			strResponse += "Sorry, you can only delete your own posts\n"
		} else {
			strResponse += "Some posts have not been deleted because they were not yours\n"
		}
	}

	if isErrorPinnedPost {
		strResponse += "Some posts have not been deleted because they were pinned in the channel\n"
	}

	if numDeletedPost > 0 {
		plural := ""
		if numDeletedPost > 1 {
			plural = "s"
		}
		strResponse += fmt.Sprintf("Successfully deleted %d post%s", numDeletedPost, plural)
	}

	if strResponse == "" {
		strResponse = "There are no posts in this channel"
	}

	p.sendEphemeralPost(&model.CommandArgs{
		ChannelId: channelID,
		UserId:    userID,
	}, strResponse)
	return nil
}
