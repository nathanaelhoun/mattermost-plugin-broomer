package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	routeDialogDeleteLast    = "/dialog/delete-last"
	routeDialogDeleteFilters = "/dialog/delete-filters"
	routeAutocompletePostID  = "/autocomplete/postid"
)

// ServeHTTP allows the plugin to implement the http.Handler interface. Requests destined for the
// /plugins/{id} path will be routed to the plugin.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case routeDialogDeleteLast:
		dialogDeleteLast(p, w, r)

	case routeDialogDeleteFilters:
		dialogDeleteWithFilters(p, w, r)

	case routeAutocompletePostID:
		if r.Method != http.MethodGet {
			p.respondAndLogErr(w, http.StatusMethodNotAllowed, errors.New("method"+r.Method+"is not allowed, must be GET"))
			return
		}
		autocompletePostID(p, w, r)

	default:
		http.NotFound(w, r)
	}
}

// respondAndLogErr log the error in the server and send a response code to the request
func (p *Plugin) respondAndLogErr(w http.ResponseWriter, code int, err error) {
	http.Error(w, err.Error(), code)
	p.API.LogError(err.Error())
}

// respondJSON turn the object into a JSON response and send it
func (p *Plugin) respondJSON(w http.ResponseWriter, obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		p.respondAndLogErr(w, http.StatusInternalServerError, errors.WithMessage(err, "failed to marshal response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		p.respondAndLogErr(w, http.StatusInternalServerError, errors.WithMessage(err, "failed to write response"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// autocompletePostID is a "fake" dynamic list that show the post is the postID given by the user
// is correct.
func autocompletePostID(p *Plugin, w http.ResponseWriter, r *http.Request) {
	// query := r.URL.Query()
	// fullUserInput := query.Get("user_input")
	// p.API.LogDebug("Received autocomplete request", "userInput", fullUserInput)
	// split := strings.Fields(fullUserInput)
	// userInput := split[len(split)-1]

	var out []model.AutocompleteListItem

	out = append(out,
		model.AutocompleteListItem{
			HelpText: "Input the post ID",
			Item:     "[postID|postURL]",
		},
	)

	// TODO add autocomplete
	// filter.AddNamedTextArgument(filterArgAfter, "Delete posts after this one", "[postID|postURL]", "", false)
	// filter.AddNamedTextArgument(filterArgBefore, "Delete posts before this one", "[postID|postURL]", "", false)

	// postID, err := transformToPostID(p, userInput, query.Get("channel_id"))
	// if err != nil {
	// 	out = append(out,
	// 		model.AutocompleteListItem{
	// 			HelpText: err.Error(),
	// 			Item:     "Error",
	// 		},
	// 	)
	// } else {
	// 	post, appErr := p.API.GetPost(postID)
	// 	if appErr != nil {
	// 		p.API.LogError("Unable to get post", "postID", postID)
	// 	} else {
	// 		poster, _ := p.API.GetUser(post.UserId)

	// 		out = append(out, model.AutocompleteListItem{
	// 			HelpText: fmt.Sprintf("Posted by @%s: \"%s\"", poster.Username, post.Message),
	// 			Item:     post.Id,
	// 		})
	// 	}
	// }

	p.respondJSON(w, out)
}

// dialogDeleteLast handles the relevant interactive dialog
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
		p.API.LogError("Failed to convert string to int. Bad request", "err", err.Error(), "request", request)
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

// dialogDeleteWithFilters handles the relevant interactive dialog
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
