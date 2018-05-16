package main

import (
	"os"

	urfave "github.com/urfave/cli"

	"wizeBlock/wizeNode/cli"
)

func main() {
	app := urfave.NewApp()
	app.Name = "WizeBlock Node"
	app.Version = "0.2"
	app.Usage = "Command-line API for WizeBlock Node"

	app.Flags = cli.GlobalFlags
	app.Commands = cli.Commands

	app.CommandNotFound = cli.CommandNotFound
	app.Before = cli.CommandBefore

	app.Run(os.Args)
}
