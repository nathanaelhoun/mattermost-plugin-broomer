package main

import (
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) dialogDeleteLast(w http.ResponseWriter, r *http.Request) {
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

	deletePinnedPost := false
	if request.Submission["deletePinnedPosts"] == true {
		deletePinnedPost = true
	}

	p.deleteLastPostsInChannel(numPostToDelete, request.ChannelId, request.UserId, deletePinnedPost)
}
