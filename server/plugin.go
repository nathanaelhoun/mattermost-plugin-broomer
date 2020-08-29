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
	botUserID, err := p.Helpers.EnsureBot(
		&model.Bot{
			Username:    "broomerbot",
			DisplayName: "Broomer",
			Description: "Bot managed by the Broomer plugin.",
		},
		plugin.IconImagePath("/assets/broom.svg"),
		plugin.ProfileImagePath("/assets/broom.png"),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to ensure bot")
	}

	_, appErr := p.API.UpdateBotActive(botUserID, true)
	if appErr != nil {
		return errors.Wrap(appErr, "Failed mark the bot as active")
	}
	p.botUserID = botUserID

	// Registering command in OnConfigurationChange()
	return nil
}

func (p *Plugin) OnDeactivate() error {
	_, appErr := p.API.UpdateBotActive(p.botUserID, false)
	return errors.Wrap(appErr, "Failed to mark the bot as inactive")
}
