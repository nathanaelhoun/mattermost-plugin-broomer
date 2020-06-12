package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

func hasAdminRights(p *Plugin, userID string) bool {
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("Unable to get user", "err", appErr)
	}

	return strings.Contains(user.Roles, "system_admin")
}

func (p *Plugin) sendEphemeralPost(args *model.CommandArgs, message string) *model.Post {
	return p.API.SendEphemeralPost(
		args.UserId,
		&model.Post{
			UserId:    p.botUserID,
			ChannelId: args.ChannelId,
			Message:   message,
		},
	)
}

func (p *Plugin) deletePosts(numPostToDelete int, channelID string, userID string) *model.AppError {
	postList, err := p.API.GetPostsForChannel(channelID, 0, numPostToDelete)
	if err != nil {
		p.API.LogError(
			"Unable to retrieve posts",
			"err", err.Error(),
		)
		return err
	}

	isError := false
	isErrorNotAdmin := false
	numDeletedPost := 0
	hasAdminRights := hasAdminRights(p, userID)
	for _, postID := range postList.Order {
		if !hasAdminRights {
			post, err := p.API.GetPost(postID)
			if err != nil {
				isError = true
				p.API.LogError(
					"Unable to get post "+postID+" informations.",
					"err", err.Error(),
				)
				continue // process next post
			}

			if post.UserId != userID {
				isErrorNotAdmin = true
				continue // process next post
			}
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
		strResponse += "An error has occurred, some post could not be deleted.\n"
	}

	if isErrorNotAdmin {
		strResponse += "Some posts have not been deleted because they were not yours.\n"
	}

	if numDeletedPost > 0 {
		plural := ""
		if numDeletedPost > 1 {
			plural = "s"
		}
		strResponse += fmt.Sprintf("Successfully deleted %d post%s.", numDeletedPost, plural)
	}

	if strResponse == "" {
		strResponse = "There are no posts in this channel."
	}

	p.sendEphemeralPost(&model.CommandArgs{
		ChannelId: channelID,
		UserId:    userID,
	}, strResponse)
	return nil
}
