package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "WizeBlock Wallet"
	app.Version = "0.2"
	app.Usage = "Command-line API for WizeBlock Wallet"

	app.Flags = GlobalFlags
	app.Commands = Commands

	app.CommandNotFound = CommandNotFound
	app.Before = CommandBefore

	app.Run(os.Args)
}
