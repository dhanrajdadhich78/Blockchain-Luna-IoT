package main

import (
	"wizeBlockchain/code"

)

func main() {
	bc := code.NewBlockchain()
	defer bc.Db.Close()

	cli := code.CLI{bc}
	cli.Run()
}