package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

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
