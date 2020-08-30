package main

import "github.com/mattermost/mattermost-server/v5/model"

// TODO All this command

type delFilters struct {
	afterThisPostID  string
	beforeThisPostID string
	fromUsersIDs     map[string]bool
}

func (p *Plugin) executeCommandFilter(options *delOptions, filters *delFilters) (*model.CommandResponse, *model.AppError) {
	p.sendEphemeralPost(options.userID, options.channelID, "To be done : This command is not yet implemented")
	return &model.CommandResponse{}, nil
}
