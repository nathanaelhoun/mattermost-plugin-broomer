package main

import (
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	commandTrigger  = "broom"
	commandHint     = "[subcommand]"
	commandHelpText = "Clean the channel by removing posts. Available commands: " + lastTrigger + ", " + filterTrigger + ", " + helpTrigger

	helpTrigger  = "help"
	helpHint     = ""
	helpHelpText = "Learn how to broom"

	messageBeginning = "Beginning housecleaning, please wait..."

	argDeletePinnedPost = "delete-pinned-posts"
	argNoConfirm        = "confirm"
)

// userError is an error that must be shown to the user, but that is not relevant to be logged by the server
type userError error

// getCommand return the slash command
// The autocompleteData is generated according to the configuration
func getCommand(conf *configuration) *model.Command {
	cmdAutocompleteData := model.NewAutocompleteData(commandTrigger, commandHint, commandHelpText)
	if conf.RestrictToSysadmins {
		cmdAutocompleteData.RoleID = "system_admin"
	}

	last := model.NewAutocompleteData(lastTrigger, lastHint, lastHelpText)
	last.AddTextArgument(last.HelpText, lastHint, "[0-9]+")
	addAllNamedTextArguments(last, conf.AskConfirm == askConfirmOptional)

	filter := model.NewAutocompleteData(filterTrigger, filterHint, filterHelpText)
	filter.AddNamedDynamicListArgument(filterArgAfter, "Delete posts after this one", routeAutocompletePostID, false)
	filter.AddNamedDynamicListArgument(filterArgBefore, "Delete posts before this one", routeAutocompletePostID, false)
	filter.AddNamedTextArgument(filterArgFrom, "Delete posts posted by a specific user", "[@username]", "@.+", false)
	addAllNamedTextArguments(filter, conf.AskConfirm == askConfirmOptional)

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

// addAllNamedTextArguments add all the flags that comes with all the commands
func addAllNamedTextArguments(cmd *model.AutocompleteData, disableConfirmDialog bool) {
	// TODO use boolean flag when then are available. See https://github.com/mattermost/mattermost-server/pull/14810
	cmd.AddNamedTextArgument(argDeletePinnedPost, "Also delete pinned posts (disabled by default)", "true", "", false)
	if disableConfirmDialog {
		cmd.AddNamedTextArgument(argNoConfirm, "Do not show confirmation dialog", "true", "", false)
	}
}

// getHelp generate a string with an help message for the user
// It uses the conf to know which flags to are enabled on the instance
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

// ExecuteCommand is called when a slash command registered by the plugin is used
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Respond "no trigger found" if the user is not authorized
	if p.getConfiguration().RestrictToSysadmins && !isSysadmin(p, args.UserId) {
		return nil, nil
	}

	subcommand, options, filters, userErr := p.parseAndCheckCommandArgs(args)
	p.API.LogDebug("Parsed command broom", "subcommand", subcommand, "options", options, "filters", filters)
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

// parseAndCheckCommandArgs returns the subcommand, the options and the filter in the command, sanitized
func (p *Plugin) parseAndCheckCommandArgs(args *model.CommandArgs) (string, *delOptions, *delFilters, userError) {
	subcommand := ""
	options := &delOptions{
		channelID:             args.ChannelId,
		userID:                args.UserId,
		triggerID:             args.TriggerId,
		numPost:               0,
		permDeleteOthersPosts: canDeleteOthersPosts(p, args.UserId, args.ChannelId),
		optDeletePinnedPosts:  false,
		optNoConfirmDialog:    false,
	}

	filters := &delFilters{
		afterThisPostID:  "",
		beforeThisPostID: "",
		fromUsersIDs:     make(map[string]bool),
	}

	split := strings.Fields(args.Command)
	if len(split) < 2 || split[1] == helpTrigger {
		return helpTrigger, nil, nil, nil
	}
	subcommand = split[1]

	for i := 2; i < len(split); i++ { // Initialize to 2 to skip '/broom subcommand'
		// This is a namedTextArgument: process the argument and its value
		if strings.HasPrefix(split[i], "--") {
			argName := split[i][2:]

			// It should have a value
			i++
			if i >= len(split) {
				return "", nil, nil, errors.Errorf(
					"Argument `--%s` should have a value. Type `/%s %s` to learn how to broom",
					argName,
					commandTrigger,
					helpTrigger,
				)
			}
			argValue := split[i]

			// TODO: improve parser with multiple users name after --from (and document it)
			argValueString, argValueBool, userErr := processNamedArgValue(p, argName, argValue, options, filters)
			if userErr != nil {
				return subcommand, nil, nil, userErr
			}

			switch argName {
			case argDeletePinnedPost:
				options.optDeletePinnedPosts = *argValueBool
			case argNoConfirm:
				options.optNoConfirmDialog = *argValueBool

			case filterArgAfter:
				filters.afterThisPostID = *argValueString
			case filterArgBefore:
				filters.beforeThisPostID = *argValueString
			case filterArgFrom:
				filters.fromUsersIDs[*argValueString] = true
			}

			continue // i has been incremented already to skip the value of the named argument
		}

		// The number of post to delete has already been defined: this argument shouldn't be here
		if options.numPost != 0 {
			return "", nil, nil, errors.Errorf("Invalid argument `%s`", split[i])
		}

		numPostToDelete64, err := strconv.ParseInt(split[i], 10, 0)
		if err != nil {
			return subcommand, nil, nil, errors.Errorf("Incorrect argument. [number-of-post] must be an integer")
		}

		if numPostToDelete64 < 1 {
			return subcommand, nil, nil, errors.Errorf("You may want to delete at least one post :wink:")
		}

		currentChannel, appErr := p.API.GetChannel(args.ChannelId)
		if appErr != nil {
			p.API.LogError("Unable to get channel statistics", "Error:", appErr)
			return subcommand, nil, nil, errors.Errorf("Error when deleting posts")
		}

		if currentChannel.TotalMsgCount < numPostToDelete64 {
			// stop the command because if numPostToDelete > currentChannel.TotalMsgCount, the plugin crashes
			return subcommand, nil, nil, errors.Errorf("Cannot delete more posts that there is in this channel")
		}

		options.numPost = int(numPostToDelete64)
	}

	if filters.afterThisPostID != "" && filters.beforeThisPostID != "" {
		// Check that there can be posts between the first and the last one
		firstPost, appErr := p.API.GetPost(filters.afterThisPostID)
		if appErr != nil {
			p.API.LogError("Unable tot get post", "appErr", appErr.ToJson())
			return subcommand, nil, nil, errors.Errorf("Error when deleting posts")
		}

		lastPost, appErr := p.API.GetPost(filters.beforeThisPostID)
		if appErr != nil {
			p.API.LogError("Unable tot get post", "appErr", appErr.ToJson())
			return subcommand, nil, nil, errors.Errorf("Error when deleting posts")
		}

		if lastPost.CreateAt < firstPost.CreateAt {
			return subcommand, nil, nil, errors.Errorf(
				"Post %s is older post %s so there is no post that are before the first and after the last",
				firstPost.Id,
				lastPost.Id,
			)
		}
	}

	// Check that subcommand, options and filters are compatible
	if subcommand == lastTrigger {
		if filters.afterThisPostID != "" || filters.beforeThisPostID != "" || len(filters.fromUsersIDs) > 0 {
			return subcommand, nil, nil, errors.Errorf(
				"Sorry, you can't use filters with `/%s %s`",
				commandTrigger,
				lastTrigger,
			)
		}
	}

	if subcommand == filterTrigger {
		if options.numPost != 0 {
			return subcommand, nil, nil, errors.Errorf(
				"Invalid argument `%d` with `/%s %s`. Please type `/%s %s` to learn how to broom",
				options.numPost,
				commandTrigger,
				filterTrigger,
				commandTrigger,
				helpTrigger,
			)
		}
	}

	// All is good!
	return subcommand, options, filters, nil
}

// Process a named arg defined for this command and check its value
func processNamedArgValue(p *Plugin, argName string, argValue string, existingOptions *delOptions, existingFilters *delFilters) (*string, *bool, userError) {
	switch argName {
	// --------------------------------------------
	case argNoConfirm:
		fallthrough
	case argDeletePinnedPost:
		if argValue == "true" || argValue == "false" {
			value := (argValue == "true")
			return nil, &value, nil
		}
		return nil, nil, errors.Errorf("Invalid value for `--%s`, `%s` should be `true` or `false`", argName, argValue)

	// --------------------------------------------
	case filterArgAfter:
		if existingFilters.afterThisPostID != "" {
			return nil, nil, errors.Errorf(
				"Argument `--%s` can only be used one. It is already defined to `%s`",
				argName,
				existingFilters.afterThisPostID,
			)
		}

		postID, userErr := transformToPostID(p, argValue, existingOptions.channelID)
		if userErr != nil {
			return nil, nil, errors.Errorf("Incorrect value for `--%s`: %s", argName, userErr.Error())
		}
		return &postID, nil, nil

	// --------------------------------------------
	case filterArgBefore:
		if existingFilters.beforeThisPostID != "" {
			return nil, nil, errors.Errorf(
				"Argument `--%s` can only be used one. It is already defined to `%s`",
				argName,
				existingFilters.beforeThisPostID,
			)
		}

		postID, userErr := transformToPostID(p, argValue, existingOptions.channelID)
		if userErr != nil {
			return nil, nil, errors.Errorf("Incorrect value for `--%s`: %s", argName, userErr.Error())
		}
		return &postID, nil, nil

	// --------------------------------------------
	case filterArgFrom:
		user, appErr := p.API.GetUserByUsername(strings.TrimLeft(argValue, "@"))
		if appErr != nil {
			// TODO change message if internal error or user unknown
			p.API.LogError("Unable to get user", "appError :", appErr.ToJson())
			return nil, nil, errors.Errorf("Invalid value for argument `--%s` : user `%s` is unknown", argName, argValue)
		}

		// TODO check is the user is in the channel

		return &user.Id, nil, nil

	// --------------------------------------------
	default:
		return nil, nil, errors.Errorf(
			"Unknown argument `--%s`. Type `/%s %s` to learn how to broom",
			argName,
			commandTrigger,
			helpTrigger,
		)
	}
}
