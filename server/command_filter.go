package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type delFilters struct {
	afterThisPostID  string
	beforeThisPostID string
	fromUsersIDs     map[string]bool
}

func (p *Plugin) executeCommandFilter(options *delOptions, filters *delFilters) (*model.CommandResponse, *model.AppError) {
	conf := p.getConfiguration()
	if conf.AskConfirm == askConfirmNever ||
		(conf.AskConfirm == askConfirmOptional && options.optNoConfirmDialog) {
		// Delete posts without confirmation dialog
		p.deletePostWithFilters(options, filters)
		return &model.CommandResponse{}, nil
	}

	// Send a confirmation dialog
	p.sendDialogDeleteFilters(options, filters)
	p.sendEphemeralPost(options.userID, options.channelID, "To be done: this is not implemented yet, sorry")
	return &model.CommandResponse{}, nil
}

func (p *Plugin) sendDialogDeleteFilters(options *delOptions, filters *delFilters) {
	// TODO
	// siteURL := ""
	// if p.API.GetConfig().ServiceSettings.SiteURL != nil {
	// 	siteURL = *p.API.GetConfig().ServiceSettings.SiteURL
	// }

	// dialog := &model.OpenDialogRequest{
	// 	URL: fmt.Sprintf("%s/plugins/%s%s", siteURL, manifest.Id, routeDialogDeleteLast),
	// 	Dialog: model.Dialog{
	// 		CallbackId: "confirmPostDeletion",
	// 		Title: fmt.Sprintf(
	// 			"Do you want to delete the last %d post%s in this channel?",
	// 			options.numPost,
	// 			getPluralChar(options.numPost),
	// 		),
	// 		SubmitLabel:    "Confirm",
	// 		NotifyOnCancel: false,
	// 		State:          "TODO",
	// 		Elements: []model.DialogElement{
	// 			{
	// 				Type:        "bool",
	// 				Name:        "deletePinnedPosts",
	// 				DisplayName: "Delete pinned posts ?",
	// 				HelpText:    "",
	// 				Default:     strconv.FormatBool(options.optDeletePinnedPosts),
	// 				Optional:    true,
	// 			},
	// 		},
	// 	},
	// }

	// if err := p.API.OpenInteractiveDialog(*dialog); err != nil {
	// 	errorMessage := "Failed to open Interactive Dialog"
	// 	p.API.LogError(errorMessage, "err", err.Error())
	// 	p.sendEphemeralPost(options.userID, options.channelID, errorMessage)
	// }
}

func (p *Plugin) deletePostWithFilters(options *delOptions, filters *delFilters) {
	p.API.LogDebug("Delete posts with this filters", "channelID", options.channelID, "filters", filters)
	// TODO

	// check permissions
	hasPermissionToDeletePost := canDeletePost(p, options.userID, options.channelID)
	if !hasPermissionToDeletePost {
		p.sendEphemeralPost(options.channelID, options.userID, "Sorry, you are not permitted to delete posts")
		return
	}

	// post in progress
	beginningPost := p.sendEphemeralPost(options.userID, options.channelID, messageBeginning)

	// Get all posts between first post and last post
	// TODO
	postList, appErr := p.API.GetPostsForChannel(options.channelID, 0, 4)
	if appErr != nil {
		p.API.LogError(
			"Unable to retrieve posts",
			"err", appErr.Error(),
		)
		return
	}
	relevantPostList := getRelevantPostList(postList)

	// filters out posts not posted by users in the filters
	postListToDelete := relevantPostList
	if len(filters.fromUsersIDs) > 0 {
		postListToDelete = model.NewPostList()

		for _, postID := range relevantPostList.Order {
			p.API.LogDebug("Filtering this post", "post", relevantPostList.Posts[postID])
			userID := relevantPostList.Posts[postID].UserId

			if filters.fromUsersIDs[userID] {
				p.API.LogDebug("This posts matches the --from filter!", "postID", postID)
				postListToDelete.AddPost(relevantPostList.Posts[postID])
				postListToDelete.AddOrder(postID)
			}
		}
	}

	result := p.deletePosts(
		postListToDelete,
		options,
	)

	beginningPost.Message = getResponseStringFromResults(result)
	p.API.UpdateEphemeralPost(options.userID, beginningPost)
}
