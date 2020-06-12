# Plugin PostManager (Beta) 

[![CircleCI branch](https://img.shields.io/circleci/project/github/nathanaelhoun/mattermost-plugin-postmanager/master.svg)](https://circleci.com/gh/nathanaelhoun/mattermost-plugin-postmanager)

This [mattermost](https://mattermost.org) plugin allow to manage posts with commands.

![Plugin screenshot](./screenshot.png)

**Supported Mattermost Server Versions: 5.2+** (command autocomplete with Mattermost 5.24+)

## Features

#### Manage posts with commands
Available commands :
  - `/postmanage delete [number-of-post]` 	Delete posts in the channel
  - `/postmanage help` 										  Display command usage

#### Use aliases
You can toggle theses aliases in **System Console > Plugins > PostManager**
  - `/clear` is an alias for `/postmanage delete`

## Installation

1. Go to the [releases page of this Github repository](https://github.com/nathanaelhoun/mattermost-plugin-postmanage/releases) and download the latest release for your Mattermost server.
2. Upload this file in the Mattermost System Console under **System Console > Plugins > Management** to install the plugin. To learn more about how to upload a plugin, [see the documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).
3. Activate the plugin at **System Console > Plugins > Management**.


## Contribution

Build the plugin with:
```
make
```