package main

import (
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	botUserID string
}

func (p *Plugin) OnActivate() error {
	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    "clearbot",
		DisplayName: "Clear Bot",
		Description: "Bot managed by the Clear plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "Failed to ensure bot")
	}
	p.botUserID = botUserID

	if err := p.API.RegisterCommand(p.getCommand()); err != nil {
		return errors.Wrap(err, "failed to register new command")
	}

	return nil
}
