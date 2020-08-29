package main

import (
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

func canDeletePost(p *Plugin, userID string, channelID string) bool {
	return p.API.HasPermissionTo(userID, model.PERMISSION_DELETE_POST) ||
		p.API.HasPermissionToChannel(userID, channelID, model.PERMISSION_DELETE_POST)
}

func canDeleteOthersPosts(p *Plugin, userID string, channelID string) bool {
	return p.API.HasPermissionTo(userID, model.PERMISSION_DELETE_OTHERS_POSTS) ||
		p.API.HasPermissionToChannel(userID, channelID, model.PERMISSION_DELETE_OTHERS_POSTS)
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
