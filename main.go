package main

import (
	//"wizeBlockchain/code"
	"flag"
	n "wizeBlockchain/network"
)

func main() {
	//cli := code.CLI{}
	//cli.Run()

	//add network
	flag.Parse()
	n.NewNode().Run()
}
