package main

import (
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
)

func dialogDeleteLast(p *Plugin, w http.ResponseWriter, r *http.Request) {
	request := model.SubmitDialogRequestFromJson(r.Body)
	if request == nil {
		p.API.LogError("failed to decode SubmitDialogRequest")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.Cancelled {
		w.WriteHeader(http.StatusOK)
		return
	}

	numPostToDelete, err := strconv.Atoi(request.State)
	if err != nil {
		p.API.LogError("Failed to convert string to int. Bad request", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	p.deleteLastPostsInChannel(&delOptions{
		channelID:             request.ChannelId,
		userID:                request.UserId,
		numPost:               numPostToDelete,
		optDeletePinnedPosts:  request.Submission["deletePinnedPost"] == "true",
		permDeleteOthersPosts: canDeleteOthersPosts(p, request.UserId, request.ChannelId),
	})
}

func dialogDeleteWithFilters(p *Plugin, w http.ResponseWriter, r *http.Request) {
	// TODO
	// request := model.SubmitDialogRequestFromJson(r.Body)
	// if request == nil {
	// 	p.API.LogError("failed to decode SubmitDialogRequest")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// if request.Cancelled {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	// numPostToDelete, err := strconv.Atoi(request.State)
	// if err != nil {
	// 	p.API.LogError("Failed to convert string to int. Bad request", "err", err.Error())
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// w.WriteHeader(http.StatusOK)

	// p.deleteLastPostsInChannel(&delOptions{
	// 	channelID:             request.ChannelId,
	// 	userID:                request.UserId,
	// 	numPost:               numPostToDelete,
	// 	optDeletePinnedPosts:  request.Submission["deletePinnedPost"] == "true",
	// 	permDeleteOthersPosts: canDeleteOthersPosts(p, request.UserId, request.ChannelId),
	// })
}
