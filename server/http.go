package main

import (
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// ServeHTTP allows the plugin to implement the http.Handler interface. Requests destined for the
// /plugins/{id} path will be routed to the plugin.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/dialog/deletion":
		p.handleDeletion(w, r)

	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleDeletion(w http.ResponseWriter, r *http.Request) {
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
		p.API.LogError("Failed to convert string to int. Bad request.", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := p.deleteLastPosts(numPostToDelete, request.ChannelId, request.UserId); err != nil {
		p.API.LogError("Failed to delete posts", "err", err.Error())
		return
	}
}
