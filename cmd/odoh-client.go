package main

import (
	"fmt"
	"github.com/chris-wood/odoh-client/commands"
	"github.com/urfave/cli"
	"os"
	"time"
)

var (
	Version = "0"
	Tag     = "0"
)

func main() {
	app := cli.App{
		Name:                   "odohclient",
		HelpName:               "Oblivious DNS over HTTPS Client Command Line Interface",
		Usage:                  "",
		UsageText:              "",
		ArgsUsage:              "",
		Version:                fmt.Sprintf("%v - %v", Version, Tag),
		Description:            "",
		Commands:               commands.Commands,
		Flags:                  nil,
		EnableBashCompletion:   false,
		HideHelp:               false,
		HideVersion:            false,
		BashComplete:           nil,
		Before:                 nil,
		After:                  nil,
		Action:                 nil,
		CommandNotFound:        nil,
		OnUsageError:           nil,
		Compiled:               time.Time{},
		Authors:                nil,
		Copyright:              "",
		Author:                 "",
		Email:                  "",
		Writer:                 nil,
		ErrWriter:              nil,
		ExitErrHandler:         nil,
		Metadata:               nil,
		ExtraInfo:              nil,
		CustomAppHelpTemplate:  "",
		UseShortOptionHandling: false,
	}
	app.Run(os.Args)
}
