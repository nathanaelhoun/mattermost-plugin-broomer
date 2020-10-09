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

	body, _ := r.GetBody()
	defer body.Close()
	p.API.LogDebug("Received autocomplete request", "body", body)
	// p.API.GetPost()

	// filter.AddNamedTextArgument(filterArgAfter, "Delete posts after this one", "[postID|postURL]", "", false)
	// filter.AddNamedTextArgument(filterArgBefore, "Delete posts before this one", "[postID|postURL]", "", false)

	out := []model.AutocompleteListItem{
		{
			HelpText: "Manually type the project's VCS repository name",
			Item:     "[repository]",
		},
	}
	// if len(projects) == 0 {
	// 	p.respondJSON(w, out)
	// 	return
	// }

	// for _, project := range projects {
	// 	out = append(out, model.AutocompleteListItem{
	// 		HelpText: project.VCSURL,
	// 		Item:     project.Reponame,
	// 	})
	// }
	p.respondJSON(w, out)
}
