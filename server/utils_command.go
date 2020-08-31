package main

import (
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

// Return the subcommand, the options and the filter in the command, sanitized
func (p *Plugin) parseAndCheckCommandArgs(args *model.CommandArgs) (string, *delOptions, *delFilters, userError) {
	subcommand := ""
	options := &delOptions{
		channelID:             args.ChannelId,
		userID:                args.UserId,
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

	for i := 1; i < len(split); i++ { // Initialize to 1 to skip '/commandTrigger'
		if i == 1 {
			subcommand = split[i]
			if subcommand == helpTrigger {
				return subcommand, nil, nil, nil
			}

			continue
		}

		if strings.HasPrefix(split[i], "--") {
			// Process the argument and its value
			argName := split[i][2:]

			// If should have a value
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

		// Number to delete
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
				"Post %s is older post %s so there is no before the first and after the last",
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
func processNamedArgValue(
	p *Plugin,
	argName string,
	argValue string,
	existingOptions *delOptions,
	existingFilters *delFilters,
) (*string, *bool, userError) {
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
	}

	// --------------------------------------------
	return nil, nil, errors.Errorf(
		"Unknown argument `--%s`. Type `/%s %s` to learn how to broom",
		argName,
		commandTrigger,
		helpTrigger,
	)
}

// Check that a postID in form of the direct ID or a link to the post is correct
// If so, returns the postID
// If incorrect, returns "" and an error
func transformToPostID(p *Plugin, postIDToParse string, channelID string) (string, error) {
	if strings.HasPrefix(postIDToParse, "http") {
		// TODO: This is a link: transform it in a postID
		return "", errors.Errorf("Sorry, links are not supported for the moment. Please use the postID")
	}

	post, appErr := p.API.GetPost(postIDToParse)
	if appErr != nil {
		// TODO change message if internal error or user unknown
		p.API.LogError("Unable to get post", "appError :", appErr.ToJson())
		return "", errors.Errorf("unknown post `%s`", postIDToParse)
	}

	if post.ChannelId != channelID {
		return "", errors.Errorf("post `%s` is not in this channel", postIDToParse)
	}

	return post.Id, nil
}
