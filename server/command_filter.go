package main

import "github.com/mattermost/mattermost-server/v5/model"

// TODO All this command

func (p *Plugin) executeCommandFilter(cmdArgs *model.CommandArgs, parsedParams []string, parsedArgs map[string]bool) (*model.CommandResponse, *model.AppError) {
	return p.respondEphemeralResponse(cmdArgs, "To be done : This command is not yet implemented"), nil
}
