package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	commandTrigger = "clear"

	optionDeletePinnedPost = "delete-pinned-posts"

	commandHelpText = "## Delete posts with /" + commandTrigger + "\n" +
		"`/clear [number-of-post]` Delete the last `[number-of-post]` posts in the current channel\n" +
		"\n" +
		"### Available options :\n" +
		" * `--" + optionDeletePinnedPost + "` Also delete pinned post (disabled by default)\n"
)

func (p *Plugin) getCommand() *model.Command {
	return &model.Command{
		Trigger:          commandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Delete posts",
		AutoCompleteHint: "[number-of-posts]",
		AutocompleteData: getAutocompleteData(),
	}
}

func getAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData(commandTrigger, "[number-of-posts]", "Delete posts in the current channel")

	command.AddTextArgument("Delete the last [number-of-post] posts in this channel", "[number-of-post]", "[0-9]+")
	command.AddNamedTextArgument(optionDeletePinnedPost, "Also delete pinned posts (disabled by default)", "true", "", false)

	return command
}

func parseArguments(args *model.CommandArgs) ([]string, map[string]bool) {
	parameters := []string{}
	options := make(map[string]bool)

	nextIsNamedTextArgumentValue := false
	namedTextArgumentName := ""

	for position, arg := range strings.Fields(args.Command) {
		if position == 0 {
			continue // skip '/commandTrigger'
		}

		if nextIsNamedTextArgumentValue {
			// NamedTextArgument should only be "true" or "false" in this plugin
			if arg == "false" {
				delete(options, namedTextArgumentName)
			}

			nextIsNamedTextArgumentValue = false
			namedTextArgumentName = ""
			continue
		}

		if strings.HasPrefix(arg, "--") {
			optionName := arg[2:]
			options[optionName] = true
			nextIsNamedTextArgumentValue = true
			namedTextArgumentName = optionName
			continue
		}

		parameters = append(parameters, arg)
	}

	return parameters, options
}

func (p *Plugin) verifyCommandDelete(parameters []string, args *model.CommandArgs) (int, *model.AppError) {
	if len(parameters) < 1 {
		p.sendEphemeralPost(args, "Please precise the [number-of-post] you want to delete")
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

	currentChannel, appErr := p.API.GetChannel(args.ChannelId)
	if appErr != nil {
		p.sendEphemeralPost(args, "Error when deleting posts")
		return 0, &model.AppError{
			Message:       "Unable to get channel statistics",
			DetailedError: appErr.DetailedError,
		}
	}
	if currentChannel.TotalMsgCount < numPostToDelete64 {
		// stop the command because if numPostToDelete > currentChannel.TotalMsgCount, the plugin crashes
		p.sendEphemeralPost(args, "Cannot delete more posts that there is in this channel")
		return 0, nil
	}

	return int(numPostToDelete64), nil
}

func (p *Plugin) askConfirmCommandDelete(numPostToDelete int, args *model.CommandArgs, deletePinnedPosts bool) (*model.CommandResponse, *model.AppError) {
	serverConfig := p.API.GetConfig()

	dialog := &model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       fmt.Sprintf("%s/plugins/%s/dialog/deletion", *serverConfig.ServiceSettings.SiteURL, manifest.Id),
		Dialog: model.Dialog{
			CallbackId:     "confirmPostDeletion",
			Title:          fmt.Sprintf("Do you want to delete the last %d posts in this channel?", numPostToDelete),
			SubmitLabel:    "Confirm",
			NotifyOnCancel: false,
			State:          strconv.Itoa(numPostToDelete),
			Elements: []model.DialogElement{
				{
					Type:        "bool",
					Name:        "deletePinnedPosts",
					DisplayName: "Delete pinned posts ?",
					HelpText:    "Pinned posts are keept by default",
					Default:     strconv.FormatBool(deletePinnedPosts),
					Optional:    true,
				},
			},
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

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	parameters, options := parseArguments(args)

	if len(parameters) < 1 {
		p.sendEphemeralPost(args, commandHelpText)
		return &model.CommandResponse{}, nil
	}

	numPostToDelete, err := p.verifyCommandDelete(parameters, args)
	if err != nil || numPostToDelete == 0 {
		return &model.CommandResponse{}, err
	}

	deletePinnedPost := options[optionDeletePinnedPost]
	return p.askConfirmCommandDelete(numPostToDelete, args, deletePinnedPost)
}
