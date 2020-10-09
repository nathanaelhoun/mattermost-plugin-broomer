package main

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func autocompletePostID(p *Plugin, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		p.respondAndLogErr(w, http.StatusMethodNotAllowed, errors.New("method"+r.Method+"is not allowed, must be GET"))
		return
	}

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
