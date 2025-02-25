package main

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	routeDialogDeleteLast = "/dialog/deletion"
)

// ServeHTTP allows the plugin to implement the http.Handler interface. Requests destined for the
// /plugins/{id} path will be routed to the plugin.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case routeDialogDeleteLast:
		p.dialogDeleteLast(w, r)

	default:
		http.NotFound(w, r)
	}
}
