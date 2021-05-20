package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	helpTrigger = "help"

	messageBeginning = "Beginning housecleaning, please wait..."

	argDeletePinnedPost = "delete-pinned-posts"
	argNoConfirm        = "confirm"
)

func getCommand(conf *configuration) *model.Command {
	const (
		command         = "broom"
		commandHint     = "[subcommand]"
		commandHelpText = "Clean the channel by removing posts. Available commands: " + lastTrigger + ", " + helpTrigger
	)

	cmdAutocompleteData := model.NewAutocompleteData(command, commandHint, commandHelpText)
	if conf.RestrictToSysadmins {
		cmdAutocompleteData.RoleID = "system_admin"
	}

	last := model.NewAutocompleteData(lastTrigger, lastHint, lastHelpText)
	last.AddTextArgument(last.HelpText, lastHint, "[0-9]+")
	addAllNamedTextArgumentsToCmd(last, conf.AskConfirm == askConfirmOptional)

	help := model.NewAutocompleteData(helpTrigger, "", "Learn how to broom")

	cmdAutocompleteData.AddCommand(last)
	cmdAutocompleteData.AddCommand(help)

	return &model.Command{
		Trigger:          command,
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
		" * `/broom " + lastTrigger + " " + lastHint + "` " + lastHelpText + "\n" +

		"\n" +
		"### Global arguments :\n" +
		" * `--" + argDeletePinnedPost + "` Also delete pinned post (disabled by default)\n"

	if conf.AskConfirm == askConfirmOptional {
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
