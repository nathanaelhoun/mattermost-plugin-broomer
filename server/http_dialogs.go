package main

import (
	"net/http"
)

func (p *Plugin) dialogDeleteLast(w http.ResponseWriter, r *http.Request) {
	// request := model.SubmitDialogRequestFromJson(r.Body)
	// if request == nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// if request.Canceled {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	// numPostToDelete, err := strconv.Atoi(request.State)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// w.WriteHeader(http.StatusOK)

	// p.deleteLastPostsInChannel(&deletionOptions{
	// 	channelID:             request.ChannelId,
	// 	userID:                request.UserId,
	// 	numPost:               numPostToDelete,
	// 	optDeletePinnedPosts:  request.Submission["deletePinnedPost"] == "true",
	// 	permDeleteOthersPosts: canDeleteOthersPosts(p, request.UserId, request.ChannelId),
	// })
}
