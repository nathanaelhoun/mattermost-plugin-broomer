package main

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	routeDialogDeleteLast    = "/dialog/delete-last"
	routeDialogDeleteFilters = "/dialog/delete-filters"
)

// ServeHTTP allows the plugin to implement the http.Handler interface. Requests destined for the
// /plugins/{id} path will be routed to the plugin.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	// TODO use mux.Router
	switch r.URL.Path {
	case routeDialogDeleteLast:
		dialogDeleteLast(p, w, r)

	case routeDialogDeleteFilters:
		dialogDeleteWithFilters(p, w, r)

	default:
		http.NotFound(w, r)
	}
}
