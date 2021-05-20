package root

import (
	_ "embed" // Need to embed manifest file
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

//go:embed plugin.json
var manifestString string

// Manifest contains the content extracted from plugin.json
var Manifest model.Manifest

func init() {
	Manifest = *model.ManifestFromJson(strings.NewReader(manifestString))
}
