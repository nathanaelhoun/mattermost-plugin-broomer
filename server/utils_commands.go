package main

import (
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

// userError defines an error made by the user (incorrect argument, etc).
// It should not be logged by the server, but sent back to the user
type userError error

type deletionOptions struct {
	channelID             string
	userID                string
	triggerID             string
	numPost               int
	optDeletePinnedPosts  bool
	optNoConfirmDialog    bool
	permDeleteOthersPosts bool
}

// Return the subcommand, the options in the command sanitized, and a userError if applicable
func (p *Plugin) parseAndCheckCommandArgs(args *model.CommandArgs) (string, *deletionOptions, userError) {
	subcommand := ""
	options := &deletionOptions{
		channelID:             args.ChannelId,
		userID:                args.UserId,
		triggerID:             args.TriggerId,
		numPost:               0,
		permDeleteOthersPosts: canDeleteOthersPosts(p, args.UserId, args.ChannelId),
		optDeletePinnedPosts:  false,
		optNoConfirmDialog:    false,
	}

	split := strings.Fields(args.Command)

	for i := 1; i < len(split); i++ { // Initialize to 1 to skip '/broom'
		if i == 1 {
			subcommand = split[i]
			if subcommand == helpTrigger {
				return subcommand, nil, nil
			}

			continue
		}

		if strings.HasPrefix(split[i], "--") {
			// Process the argument and its value
			argName := split[i][2:]

			// If should have a value
			i++
			if i >= len(split) {
				return "", nil, errors.Errorf(
					"Argument `--%s` should have a value. Type `/broom %s` to learn how to broom",
					argName, helpTrigger,
				)
			}
			argValue := split[i]

			_, argValueBool, userErr := processNamedArgValue(p, argName, argValue, options)
			if userErr != nil {
				return subcommand, nil, userErr
			}

			switch argName {
			case argDeletePinnedPost:
				options.optDeletePinnedPosts = *argValueBool
			case argNoConfirm:
				options.optNoConfirmDialog = *argValueBool
			}

			continue // i has been incremented already to skip the value of the named argument
		}

		// Number of post to delete
		if options.numPost != 0 {
			return "", nil, errors.Errorf("Invalid argument `%s`", split[i])
		}

		numPostToDelete64, err := strconv.ParseInt(split[i], 10, 0)
		if err != nil {
			return subcommand, nil, errors.Errorf("Incorrect argument. [number-of-post] must be an integer")
		}

		if numPostToDelete64 < 1 {
			return subcommand, nil, errors.Errorf("You may want to delete at least one post :wink:")
		}

		currentChannel, appErr := p.API.GetChannel(args.ChannelId)
		if appErr != nil {
			p.API.LogError("Unable to get channel statistics", "Error:", appErr)
			return subcommand, nil, errors.Errorf("Error when deleting posts")
		}

		if currentChannel.TotalMsgCount < numPostToDelete64 {
			// stop the command because if numPostToDelete > currentChannel.TotalMsgCount, the plugin crashes
			return subcommand, nil, errors.Errorf("Cannot delete more posts that there is in this channel")
		}

		options.numPost = int(numPostToDelete64)
	}

	// All is good!
	return subcommand, options, nil
}

// Process a named arg defined for this command and check its value
func processNamedArgValue(p *Plugin, argName string, argValue string, existingOptions *deletionOptions) (*string, *bool, userError) {
	switch argName {
	// --------------------------------------------
	case argDeletePinnedPost:
		if argValue == "true" || argValue == "false" {
			value := argValue == "true"
			return nil, &value, nil
		}
		return nil, nil, errors.Errorf("Invalid value for `--%s`, `%s` should be `true` or `false`", argName, argValue)

	// --------------------------------------------
	case argNoConfirm:
		if argValue == "true" || argValue == "false" {
			value := argValue == "true"
			return nil, &value, nil
		}
		return nil, nil, errors.Errorf("Invalid value for `--%s`, `%s` should be `true` or `false`", argName, argValue)
	}

	// --------------------------------------------
	return nil, nil, errors.Errorf("Unknown argument `--%s`. Type `/broom %s` to learn how to broom", argName, helpTrigger)
}

// addAllNamedTextArgumentsToCmd add the common named arguments autocompletion to the given command
func addAllNamedTextArgumentsToCmd(cmd *model.AutocompleteData, isDisableConfirmDialogAutocompleteEnabled bool) {
	cmd.AddNamedTextArgument(argDeletePinnedPost, "Also delete pinned posts (disabled by default)", "true", "", false)
	if isDisableConfirmDialogAutocompleteEnabled {
		cmd.AddNamedTextArgument(argNoConfirm, "Do not show confirmation dialog", "true", "", false)
	}
}
