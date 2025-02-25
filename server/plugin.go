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

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	if p.API.GetConfig().ServiceSettings.SiteURL == nil {
		return errors.Errorf("SiteURL is not configured. Please head to the System Console > Environment > Web Server > Site URL")
	}

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

	p.botUserID = botUserID

	// Registering command in OnConfigurationChange()
	return nil
}
