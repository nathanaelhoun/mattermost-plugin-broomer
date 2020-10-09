package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	commandTrigger  = "broom"
	commandHint     = "[subcommand]"
	commandHelpText = "Clean the channel by removing posts. Available commands: " + lastTrigger + ", " + filterTrigger + ", " + helpTrigger

	lastTrigger  = "last"
	lastHint     = "[number-of-posts]"
	lastHelpText = "Delete the last [number-of-posts] posts of the channel"

	filterTrigger   = "filter"
	filterHint      = "[--from|--after|--before]"
	filterHelpText  = "To Be Done : Clean this channel with filters"
	filterArgAfter  = "after"
	filterArgBefore = "before"
	filterArgFrom   = "from"

	helpTrigger  = "help"
	helpHint     = ""
	helpHelpText = "Learn how to broom"

	messageBeginning = "Beginning housecleaning, please wait..."

	argDeletePinnedPost = "delete-pinned-posts"
	argNoConfirm        = "confirm"
)

// This type define an error made by the user (incorrect argument, etc).
// It should not be logged by the server but sent back to the user
type userError error

func addAllNamedTextArgumentsToCmd(cmd *model.AutocompleteData, disableConfirmDialog bool) {
	cmd.AddNamedTextArgument(argDeletePinnedPost, "Also delete pinned posts (disabled by default)", "true", "", false)
	if disableConfirmDialog {
		cmd.AddNamedTextArgument(argNoConfirm, "Do not show confirmation dialog", "true", "", false)
	}
}

func getCommand(conf *configuration) *model.Command {
	cmdAutocompleteData := model.NewAutocompleteData(commandTrigger, commandHint, commandHelpText)
	if conf.RestrictToSysadmins {
		cmdAutocompleteData.RoleID = "system_admin"
	}

	last := model.NewAutocompleteData(lastTrigger, lastHint, lastHelpText)
	last.AddTextArgument(last.HelpText, lastHint, "[0-9]+")
	addAllNamedTextArgumentsToCmd(last, conf.AskConfirm == askConfirmOptional)

	filter := model.NewAutocompleteData(filterTrigger, filterHint, filterHelpText)
	filter.AddNamedDynamicListArgument(filterArgAfter, "Delete posts after this one", routeAutocompletePostID, false)
	filter.AddNamedDynamicListArgument(filterArgBefore, "Delete posts before this one", routeAutocompletePostID, false)
	filter.AddNamedTextArgument(filterArgFrom, "Delete posts posted by a specific user", "[@username]", "@.+", false)
	addAllNamedTextArgumentsToCmd(filter, conf.AskConfirm == askConfirmOptional)

	help := model.NewAutocompleteData(helpTrigger, helpHint, helpHelpText)

	cmdAutocompleteData.AddCommand(last)
	cmdAutocompleteData.AddCommand(filter)
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
		" * `/" + commandTrigger + " " + filterTrigger + " " + filterHint + "` " + filterHelpText + "\n" +
		"     * `--" + filterArgAfter + " [postID|postURL]` Delete posts after this one\n" +
		"     * `--" + filterArgBefore + " [postID|postURL]` Delete posts before this one\n" +
		"     * `--" + filterArgFrom + " [@username]` Delete posts posted by a specific user\n" +
		// TODO : explains arguments furthers
		// TODO : tell about dialogs UI
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

	subcommand, options, filters, userErr := p.parseAndCheckCommandArgs(args)
	p.API.LogDebug("Parsed command", "subcommand", subcommand, "options", options, "filters", filters)
	if userErr != nil {
		return p.respondEphemeralResponse(args, userErr.Error()), nil
	}

	switch subcommand {
	case lastTrigger:
		return p.executeCommandLast(options)

	case filterTrigger:
		return p.executeCommandFilter(options, filters)

	case helpTrigger:
		fallthrough
	default:
		return p.respondEphemeralResponse(args, getHelp(p.getConfiguration())), nil
	}
}
