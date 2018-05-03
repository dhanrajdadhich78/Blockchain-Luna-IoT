package main

import (
	"os"

	"github.com/urfave/cli"

	command "wizeBlock/wizeNode/app"
)

func main() {
	app := cli.NewApp()
	app.Name = "WizeBlock Node"
	app.Version = "0.2"
	app.Usage = "Command-line API for WizeBlock Node"

	app.Flags = command.GlobalFlags
	app.Commands = command.Commands

	app.CommandNotFound = command.CommandNotFound
	app.Before = command.CommandBefore

	app.Run(os.Args)
}
