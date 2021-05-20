package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	commandTrigger  = "broom"
	commandHint     = "[subcommand]"
	commandHelpText = "Clean the channel by removing posts. Available commands: " + lastTrigger + ", " + helpTrigger

	lastTrigger  = "last"
	lastHint     = "[number-of-posts]"
	lastHelpText = "Delete the last [number-of-posts] posts of the channel"

	helpTrigger  = "help"
	helpHint     = ""
	helpHelpText = "Learn how to broom"

	messageBeginning = "Beginning housecleaning, please wait..."

	argDeletePinnedPost = "delete-pinned-posts"
	argNoConfirm        = "confirm"
)

func getCommand(conf *configuration) *model.Command {
	cmdAutocompleteData := model.NewAutocompleteData(commandTrigger, commandHint, commandHelpText)
	if conf.RestrictToSysadmins {
		cmdAutocompleteData.RoleID = "system_admin"
	}

	last := model.NewAutocompleteData(lastTrigger, lastHint, lastHelpText)
	last.AddTextArgument(last.HelpText, lastHint, "[0-9]+")
	addAllNamedTextArgumentsToCmd(last, conf.AskConfirm == AskConfirmOptional)

	help := model.NewAutocompleteData(helpTrigger, helpHint, helpHelpText)

	cmdAutocompleteData.AddCommand(last)
	cmdAutocompleteData.AddCommand(help)

	return &model.Command{
		Trigger:          commandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: commandHelpText,
		AutoCompleteHint: commandHint,
		AutocompleteData: cmdAutocompleteData,
	}
}

func getHelp(conf *configuration) string {
	helpStr := "## Broomer Plugin\n" +
		"Easily clean the current channel with this magic broom.\n" +
		"\n" +
		" * `/" + commandTrigger + " " + lastTrigger + " " + lastHint + "` " + lastHelpText + "\n" +

		"\n" +
		"### Global arguments :\n" +
		" * `--" + argDeletePinnedPost + "` Also delete pinned post (disabled by default)\n"

	if conf.AskConfirm == AskConfirmOptional {
		helpStr += " * `--" + argNoConfirm + "` Do not show confirmation dialog\n"
	}

	return helpStr
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Respond "no trigger found" if the user is not authorized
	if p.getConfiguration().RestrictToSysadmins && !isSysadmin(p, args.UserId) {
		return nil, nil
	}

	subcommand, options, userErr := p.parseAndCheckCommandArgs(args)
	if userErr != nil {
		return p.respondEphemeralResponse(args, userErr.Error()), nil
	}

	switch subcommand {
	case lastTrigger:
		return p.executeCommandLast(options)

	case helpTrigger:
		fallthrough
	default:
		return p.respondEphemeralResponse(args, getHelp(p.getConfiguration())), nil
	}
}
