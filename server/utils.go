package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

// Checks if the user has sysadmin permission
func isSysadmin(p *Plugin, userID string) bool {
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("Unable to get user", "err", appErr)
		return false
	}

	return strings.Contains(user.Roles, "system_admin")
}

// Checks if the user has the "delete_post" permission
func canDeletePost(p *Plugin, userID string, channelID string) bool {
	return p.API.HasPermissionTo(userID, model.PERMISSION_DELETE_POST) ||
		p.API.HasPermissionToChannel(userID, channelID, model.PERMISSION_DELETE_POST)
}

// Checks if the user has the "delete_others_posts" permission
func canDeleteOthersPosts(p *Plugin, userID string, channelID string) bool {
	return p.API.HasPermissionTo(userID, model.PERMISSION_DELETE_OTHERS_POSTS) ||
		p.API.HasPermissionToChannel(userID, channelID, model.PERMISSION_DELETE_OTHERS_POSTS)
}

// Returns "s" if the given number is > 1
func getPluralChar(number int) string {
	if 1 < number {
		return "s"
	}

	return ""
}

// Simplified version of SendEphemeralPost, send to the userID defined
func (p *Plugin) sendEphemeralPost(userID string, channelID string, message string) *model.Post {
	return p.API.SendEphemeralPost(
		userID,
		&model.Post{UserId: p.botUserID, ChannelId: channelID, Message: message},
	)
}

// Wrapper of p.sendEphemeralPost() to one-line the return statements when a *model.CommandResponse is expected
func (p *Plugin) respondEphemeralResponse(args *model.CommandArgs, message string) *model.CommandResponse {
	_ = p.sendEphemeralPost(args.UserId, args.ChannelId, message)
	return &model.CommandResponse{}
}

// Tells if the plugin should has for the confirmation of deletion
func (p *Plugin) shouldConfirmDeletion(optNoConfirmDialog bool) bool {
	conf := p.getConfiguration()

	if conf.AskConfirm == askConfirmNever {
		return false
	}

	if conf.AskConfirm == askConfirmOptional && optNoConfirmDialog {
		return false
	}

	return true
}
