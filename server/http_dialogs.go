package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *Plugin) dialogDeleteLast(w http.ResponseWriter, r *http.Request) {
	var request *model.SubmitDialogRequest
	decodeErr := json.NewDecoder(r.Body).Decode(&request)
	if decodeErr != nil || request == nil {
		p.API.LogWarn("failed to decode SubmitDialogRequest")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	//nolint:misspell
	if request.Cancelled {
		w.WriteHeader(http.StatusOK)
		return
	}

	numPostToDelete, err := strconv.Atoi(request.State)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	p.deleteLastPostsInChannel(&deletionOptions{
		channelID:             request.ChannelId,
		userID:                request.UserId,
		numPost:               numPostToDelete,
		optDeletePinnedPosts:  request.Submission["deletePinnedPost"] == "true",
		permDeleteOthersPosts: canDeleteOthersPosts(p, request.UserId, request.ChannelId),
	})
}
