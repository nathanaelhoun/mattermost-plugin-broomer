package main

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	lastTrigger  = "last"
	lastHint     = "[number-of-posts]"
	lastHelpText = "Delete the last [number-of-posts] posts of the channel"
)

func (p *Plugin) executeCommandLast(options *deletionOptions) (*model.CommandResponse, *model.AppError) {
	conf := p.getConfiguration()

	if conf.AskConfirm == askConfirmNever ||
		(conf.AskConfirm == askConfirmOptional && options.optNoConfirmDialog) {
		// Delete posts without confirmation dialog
		p.deleteLastPostsInChannel(options)
		return &model.CommandResponse{}, nil
	}

	// Send a confirmation dialog
	p.sendDialogDeleteLast(options)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) sendDialogDeleteLast(options *deletionOptions) {
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL

	dialog := &model.OpenDialogRequest{
		TriggerId: options.triggerID,
		URL:       fmt.Sprintf("%s/plugins/%s%s", *siteURL, manifest.Id, routeDialogDeleteLast),
		Dialog: model.Dialog{
			CallbackId: "confirmPostDeletion",
			Title: fmt.Sprintf(
				"Do you want to delete the last %d post%s in this channel?",
				options.numPost, getPluralChar(options.numPost),
			),
			SubmitLabel:    "Confirm",
			NotifyOnCancel: false,
			State:          strconv.Itoa(options.numPost),
			Elements: []model.DialogElement{
				{
					Type:        "bool",
					Name:        "deletePinnedPosts",
					DisplayName: "Delete pinned posts?",
					HelpText:    "",
					Default:     strconv.FormatBool(options.optDeletePinnedPosts),
					Optional:    true,
				},
			},
		},
	}

	if err := p.API.OpenInteractiveDialog(*dialog); err != nil {
		p.API.LogError("Failed to open Interactive Dialog", "err", err)
		p.sendEphemeralPost(options.userID, options.channelID, "Failed to open Interactive Dialog")
	}
}

func (p *Plugin) deleteLastPostsInChannel(options *deletionOptions) {
	hasPermissionToDeletePost := canDeletePost(p, options.userID, options.channelID)
	if !hasPermissionToDeletePost {
		p.sendEphemeralPost(options.channelID, options.userID, "Sorry, you are not permitted to delete posts")
		return
	}

	postList, appErr := p.API.GetPostsForChannel(options.channelID, 0, options.numPost)
	if appErr != nil {
		p.API.LogError("Unable to retrieve posts", "appErr", appErr)
		p.sendEphemeralPost(options.channelID, options.userID, "Error when deleting posts")
		return
	}

	beginningPost := p.sendEphemeralPost(options.userID, options.channelID, messageBeginning)

	postListToDelete := getRelevantPostList(postList)
	result := p.deletePosts(postListToDelete, options)

	p.API.DeleteEphemeralPost(options.userID, beginningPost.Id)
	p.sendEphemeralPost(options.userID, options.channelID, result.String())
}
