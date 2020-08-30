package main

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func (p *Plugin) executeCommandLast(cmdArgs *model.CommandArgs, parsedParams []string, parsedArgs map[string]bool) (*model.CommandResponse, *model.AppError) {
	numPostToDelete, userErr := p.checkCommandLast(parsedParams, cmdArgs.ChannelId)
	if userErr != nil {
		return p.respondEphemeralResponse(cmdArgs, userErr.Error()), nil
	}

	deletePinnedPost := parsedArgs[argDeletePinnedPost]

	conf := p.getConfiguration()

	if conf.AskConfirm == AskConfirmNever ||
		(conf.AskConfirm == AskConfirmOptional && parsedArgs[argNoConfirm]) {
		// Delete posts without confirmation dialog
		p.deleteLastPostsInChannel(numPostToDelete, cmdArgs.ChannelId, cmdArgs.UserId, deletePinnedPost)
		return &model.CommandResponse{}, nil
	}

	// Send a confirmation dialog
	p.sendDialogDeleteLast(cmdArgs, numPostToDelete, deletePinnedPost)
	return &model.CommandResponse{}, nil
}

// Check the command and return an userError if the input is not correct
// Return the number of posts to delete if correct
func (p *Plugin) checkCommandLast(parsedParams []string, channelID string) (int, userError) {
	if len(parsedParams) < 1 {
		return 0, errors.Errorf("Please precise the [number-of-post] you want to delete")
	}

	numPostToDelete64, err := strconv.ParseInt(parsedParams[0], 10, 0)
	if err != nil {
		return 0, errors.Errorf("Incorrect argument. [number-of-post] must be an integer")
	}

	if numPostToDelete64 < 1 {
		return 0, errors.Errorf("You may want to delete at least one post :wink:")
	}

	currentChannel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		p.API.LogError("Unable to get channel statistics", "Error:", appErr)
		return 0, errors.Errorf("Error when deleting posts")
	}

	if currentChannel.TotalMsgCount < numPostToDelete64 {
		// stop the command because if numPostToDelete > currentChannel.TotalMsgCount, the plugin crashes
		return 0, errors.Errorf("Cannot delete more posts that there is in this channel")
	}

	return int(numPostToDelete64), nil
}

func (p *Plugin) sendDialogDeleteLast(cmdArgs *model.CommandArgs, numPostToDelete int, deletePinnedPosts bool) {
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL

	dialog := &model.OpenDialogRequest{
		TriggerId: cmdArgs.TriggerId,
		URL:       fmt.Sprintf("%s/plugins/%s%s", *siteURL, manifest.Id, routeDialogDeleteLast),
		Dialog: model.Dialog{
			CallbackId: "confirmPostDeletion",
			Title: fmt.Sprintf(
				"Do you want to delete the last %d post%s in this channel?",
				numPostToDelete,
				getPluralChar(numPostToDelete),
			),
			SubmitLabel:    "Confirm",
			NotifyOnCancel: false,
			State:          strconv.Itoa(numPostToDelete),
			Elements: []model.DialogElement{
				{
					Type:        "bool",
					Name:        "deletePinnedPosts",
					DisplayName: "Delete pinned posts ?",
					HelpText:    "",
					Default:     strconv.FormatBool(deletePinnedPosts),
					Optional:    true,
				},
			},
		},
	}

	if err := p.API.OpenInteractiveDialog(*dialog); err != nil {
		errorMessage := "Failed to open Interactive Dialog"
		p.API.LogError(errorMessage, "err", err.Error())
		p.sendEphemeralPost(cmdArgs.UserId, cmdArgs.ChannelId, errorMessage)
	}
}

func (p *Plugin) deleteLastPostsInChannel(numPostToDelete int, channelID string, userID string, deletePinnedPosts bool) {
	postList, appErr := p.API.GetPostsForChannel(channelID, 0, numPostToDelete)
	if appErr != nil {
		p.API.LogError(
			"Unable to retrieve posts",
			"err", appErr.Error(),
		)
		return
	}

	hasPermissionToDeletePost := canDeletePost(p, userID, channelID)
	if !hasPermissionToDeletePost {
		p.sendEphemeralPost(channelID, userID, "Sorry, you are not permitted to delete posts")
		return
	}

	beginningPost := p.sendEphemeralPost(userID, channelID, messageBeginning)

	postListToDelete := getRelevantPostMap(postList)

	result := p.deletePostsAndTellUser(
		&model.PostList{
			Order: postList.Order,
			Posts: postListToDelete,
		},
		&deletePostOptions{
			claimerID:         userID,
			channelID:         channelID,
			deleteOthersPosts: canDeleteOthersPosts(p, userID, channelID),
			deletePinnedPosts: deletePinnedPosts,
		},
	)

	p.API.DeleteEphemeralPost(userID, beginningPost.Id)
	p.sendEphemeralPost(userID, channelID, getResponseStringFromResults(result))
}
