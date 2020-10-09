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

func (p *Plugin) executeCommandLast(options *delOptions) (*model.CommandResponse, *model.AppError) {
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

func (p *Plugin) sendDialogDeleteLast(options *delOptions) {
	siteURL := ""
	if p.API.GetConfig().ServiceSettings.SiteURL != nil {
		siteURL = *p.API.GetConfig().ServiceSettings.SiteURL
	}

	dialog := &model.OpenDialogRequest{
		TriggerId: options.triggerID,
		URL:       fmt.Sprintf("%s/plugins/%s%s", siteURL, manifest.Id, routeDialogDeleteLast),
		Dialog: model.Dialog{
			CallbackId: "confirmPostDeletion",
			Title: fmt.Sprintf(
				"Do you want to delete the last %d post%s in this channel?",
				options.numPost,
				getPluralChar(options.numPost),
			),
			SubmitLabel:    "Confirm",
			NotifyOnCancel: false,
			State:          strconv.Itoa(options.numPost),
			Elements: []model.DialogElement{
				{
					Type:        "bool",
					Name:        "deletePinnedPosts",
					DisplayName: "Delete pinned posts ?",
					HelpText:    "",
					Default:     strconv.FormatBool(options.optDeletePinnedPosts),
					Optional:    true,
				},
			},
		},
	}

	if err := p.API.OpenInteractiveDialog(*dialog); err != nil {
		errorMessage := "Failed to open Interactive Dialog"
		p.API.LogError(errorMessage, "err", err.Error())
		p.sendEphemeralPost(options.userID, options.channelID, errorMessage)
	}
}

func (p *Plugin) deleteLastPostsInChannel(options *delOptions) {
	p.API.LogDebug("Delete posts last posts", "channelID", options.channelID, "options", options)

	postList, appErr := p.API.GetPostsForChannel(options.channelID, 0, options.numPost)
	if appErr != nil {
		p.API.LogError(
			"Unable to retrieve posts",
			"err", appErr.Error(),
		)
		return
	}

	hasPermissionToDeletePost := canDeletePost(p, options.userID, options.channelID)
	if !hasPermissionToDeletePost {
		p.sendEphemeralPost(options.channelID, options.userID, "Sorry, you are not permitted to delete posts")
		return
	}

	beginningPost := p.sendEphemeralPost(options.userID, options.channelID, messageBeginning)

	postListToDelete := getRelevantPostList(postList)

	result := p.batchDeletePosts(
		postListToDelete,
		options,
	)

	beginningPost.Message = result.String()
	p.API.UpdateEphemeralPost(options.userID, beginningPost)
}
