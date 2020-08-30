package main

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) executeCommandLast(options *delOptions) (*model.CommandResponse, *model.AppError) {
	conf := p.getConfiguration()
	if conf.AskConfirm == AskConfirmNever ||
		(conf.AskConfirm == AskConfirmOptional && options.optNoConfirmDialog) {
		// Delete posts without confirmation dialog
		p.deleteLastPostsInChannel(options)
		return &model.CommandResponse{}, nil
	}

	// Send a confirmation dialog
	p.sendDialogDeleteLast(options)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) sendDialogDeleteLast(options *delOptions) {
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL

	dialog := &model.OpenDialogRequest{
		TriggerId: options.userID,
		URL:       fmt.Sprintf("%s/plugins/%s%s", *siteURL, manifest.Id, routeDialogDeleteLast),
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

	postListToDelete := getRelevantPostMap(postList)

	result := p.deletePostsAndTellUser(
		&model.PostList{
			Order: postList.Order,
			Posts: postListToDelete,
		},
		options,
	)

	p.API.DeleteEphemeralPost(options.userID, beginningPost.Id)
	p.sendEphemeralPost(options.userID, options.channelID, getResponseStringFromResults(result))
}
