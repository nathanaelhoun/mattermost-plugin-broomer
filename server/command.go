package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	mainCommand        = "postmanage"
	subcommandDelete   = "delete"
	deleteAliasCommand = "clear"

	commandHelpText = "**Manage posts with commands.**\n" +
		"Available commands:\n" +
		"- `" + subcommandDelete + " [number-of-post]` 	Delete posts in the channel.\n" +
		"- `help` 										Display usage."
)

func (p *Plugin) getMainCommand() *model.Command {
	return &model.Command{
		Trigger:          mainCommand,
		AutoComplete:     true,
		AutoCompleteDesc: "Manage posts. Available commands : `delete`",
		AutoCompleteHint: "[command]",
		AutocompleteData: getAutocompleteData(),
	}
}

func getAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData(mainCommand, "[command]", "Manage posts. Available commands : delete")

	delete := model.NewAutocompleteData(subcommandDelete, "[number-of-post]", "Delete the last [number-of-post] posts in this channel.")
	delete.AddTextArgument("Delete the last [number-of-post] posts in this channel.", "[number-of-post]", "[0-9]+")
	command.AddCommand(delete)

	return command
}

func (p *Plugin) getDeleteAliasCommand() *model.Command {
	return &model.Command{
		Trigger:          deleteAliasCommand,
		AutoComplete:     true,
		AutoCompleteDesc: "Delete the last [number-of-post] posts in this channel. Alias for `/postmanage delete`",
		AutoCompleteHint: "[number-of-post]",
	}
}

func (p *Plugin) verifyCommandDelete(parameters []string, args *model.CommandArgs) (int, *model.AppError) {
	if len(parameters) < 1 {
		p.sendEphemeralPost(args, "Please precise the [number-of-post] you want to delete.")
		return 0, nil
	}

	numPostToDelete64, err := strconv.ParseInt(parameters[0], 10, 0)
	if err != nil {
		p.sendEphemeralPost(args, "Incorrect argument. [number-of-post] must be an integer")
		return 0, nil
	}

	if numPostToDelete64 < 1 {
		p.sendEphemeralPost(args, "Invalid number of posts")
		return 0, nil
	}

	currentChannel, err2 := p.API.GetChannel(args.ChannelId)
	if err2 != nil {
		// stop the command because if numPostToDelete > currentChannel.TotalMsgCount, the plugin crashes
		p.sendEphemeralPost(args, "Error when deleting posts.")
		return 0, &model.AppError{
			Message:       "Unable to get channel statistics",
			DetailedError: err2.DetailedError,
		}
	}
	if currentChannel.TotalMsgCount < numPostToDelete64 {
		p.sendEphemeralPost(args, "Cannot delete more posts that there is in this channel.")
		return 0, nil
	}

	return int(numPostToDelete64), nil
}

func (p *Plugin) askConfirmCommandDelete(numPostToDelete int, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	serverConfig := p.API.GetConfig()

	dialog := &model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       fmt.Sprintf("%s/plugins/%s/dialog/deletion", *serverConfig.ServiceSettings.SiteURL, manifest.Id),
		Dialog: model.Dialog{
			CallbackId:     "confirmPostDeletion",
			Title:          fmt.Sprintf("Do you really want to delete the last %d posts in this channel?", numPostToDelete),
			SubmitLabel:    "Confirm",
			NotifyOnCancel: false,
			State:          strconv.Itoa(numPostToDelete),
		},
	}

	if err := p.API.OpenInteractiveDialog(*dialog); err != nil {
		errorMessage := "Failed to open Interactive Dialog"
		p.API.LogError(errorMessage, "err", err.Error())
		p.sendEphemeralPost(args, errorMessage)
		return &model.CommandResponse{}, err
	}

	return &model.CommandResponse{}, nil
}

func parseCommand(args *model.CommandArgs) (string, []string) {
	split := strings.Fields(args.Command)

	// handle aliases
	switch strings.Trim(split[0], "/") {
	case deleteAliasCommand:
		return subcommandDelete, split[1:]

	case mainCommand:
		fallthrough
	default:
		action := ""
		if len(split) > 1 {
			action = split[1]
		}

		parameters := []string{}
		if len(split) > 2 {
			parameters = split[2:]
		}

		return action, parameters
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	action, parameters := parseCommand(args)

	switch action {
	case subcommandDelete:
		numPostToDelete, err := p.verifyCommandDelete(parameters, args)
		if err != nil || numPostToDelete == 0 {
			return &model.CommandResponse{}, err
		}

		return p.askConfirmCommandDelete(numPostToDelete, args)

	case "help":
		fallthrough
	default:
		p.sendEphemeralPost(args, commandHelpText)
		return &model.CommandResponse{}, nil
	}
}
