package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	commandTrigger  = "broom"
	commandHint     = "[subcommand]"
	commandHelpText = "Clean the channel by removing posts. Available commands: " + lastTrigger + ", " + filterTrigger + ", " + helpTrigger

	lastTrigger  = "last"
	lastHint     = "[number-of-posts]"
	lastHelpText = "Delete the last [number-of-posts] posts of the channel"

	filterTrigger   = "filter"
	filterHint      = "[--from]"
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
// It should not be logged but sent back to the user
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
	addAllNamedTextArgumentsToCmd(last, conf.AskConfirm == AskConfirmOptional)

	filter := model.NewAutocompleteData(filterTrigger, filterHint, filterHelpText)
	filter.AddNamedTextArgument(filterArgAfter, "Delete posts after this one", "[postID|postURL]", "", false)
	filter.AddNamedTextArgument(filterArgBefore, "Delete posts before this one", "[postID|postURL]", "", false)
	filter.AddNamedTextArgument(filterArgFrom, "Delete posts posted by a specific user", "[@username]", "@.+", false)
	addAllNamedTextArgumentsToCmd(filter, conf.AskConfirm == AskConfirmOptional)

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
		"     * `--" + filterArgAfter + " [postID|postURL]` Delete posts before this one\n" +
		"     * `--" + filterArgAfter + " [@username]` Delete posts posted by a specific user\n" +
		// TODO : explains arguments furthers
		"\n" +
		"### Global arguments :\n" +
		" * `--" + argDeletePinnedPost + "` Also delete pinned post (disabled by default)\n"

	if conf.AskConfirm == AskConfirmOptional {
		helpStr += " * `--" + argNoConfirm + "` Do not show confirmation dialog\n"
	}

	return helpStr
}

func parseCommandArgs(args *model.CommandArgs) (string, []string, map[string]bool, userError) {
	subcommand := ""
	parameters := []string{}
	namedArgs := make(map[string]bool)

	nextIsNamedTextArgumentValue := false
	namedTextArgumentName := ""

	for i, commandArg := range strings.Fields(args.Command) {
		if i == 0 {
			continue // skip '/commandTrigger'
		}

		if i == 1 {
			subcommand = commandArg
			continue
		}

		if nextIsNamedTextArgumentValue {
			// NamedTextArgument should only be "true" or "false" in this plugin
			switch commandArg {
			case "false":
				delete(namedArgs, namedTextArgumentName)
			case "true":
				break
			default:
				return "", nil, nil, errors.Errorf("Invalid value for argument `--%s`, must be `true` or `false`.", namedTextArgumentName)
			}

			nextIsNamedTextArgumentValue = false
			namedTextArgumentName = ""
			continue
		}

		if strings.HasPrefix(commandArg, "--") {
			optionName := commandArg[2:]
			namedArgs[optionName] = true
			nextIsNamedTextArgumentValue = true
			namedTextArgumentName = optionName
			continue
		}

		parameters = append(parameters, commandArg)
	}

	if nextIsNamedTextArgumentValue {
		return "", nil, nil, errors.Errorf("Invalid value for argument `--%s`, must be `true` or `false`.", namedTextArgumentName)
	}

	return subcommand, parameters, namedArgs, nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Respond "no trigger found" if the user is not authorized
	if p.getConfiguration().RestrictToSysadmins && !isSysadmin(p, args.UserId) {
		return nil, nil
	}

	subcommand, parameters, options, userErr := parseCommandArgs(args)
	if userErr != nil {
		return p.respondEphemeralResponse(args, userErr.Error()), nil
	}

	switch subcommand {
	case lastTrigger:
		return p.executeCommandLast(args, parameters, options)

	case filterTrigger:
		return p.executeCommandFilter(args, parameters, options)

	case helpTrigger:
		fallthrough
	default:
		return p.respondEphemeralResponse(args, getHelp(p.getConfiguration())), nil
	}
}
