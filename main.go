package main

import (
	a "wizeBlockchain/app"
	//"flag"
	//n "wizeBlockchain/network"
	//s "wizeBlockchain/blockchain/services"
	//"github.com/grrrben/golog"
	//"path/filepath"
	//"os"
	//"fmt"
	//"strconv"
	//b "wizeBlockchain/blockchain"
)

//var ClientPort uint16
//var ClientName *string

func main() {
	cli := a.CLI{}
	cli.Run()

	//add network
	//flag.Parse()
	//n.NewNode().Run()

	//prt := flag.String("p", "8000", "Port on which the app will run, defaults to 8000")
	//ClientName = flag.String("name", "0", "Set a name for the client")
	//flag.Parse()
	//
	//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	golog.Fatalf("Could not set a logdir. Msg %s", err)
	//}
	//
	//golog.SetLogDir(fmt.Sprintf("%s/log", dir))
	//
	//u, err := strconv.ParseUint(*prt, 10, 16) // always gives an uint64...
	//if err != nil {
	//	golog.Errorf("Unable to cast Port to uint: %s", err)
	//}
	//// different Clients can have different ports,
	//// used to connect multiple Clients in debug.
	//ClientPort = uint16(u)
	////s.MakeAddress()
	//a := b.App{}
	//a.ClientName = ClientName
	//a.ClientPort = ClientPort
	//a.Initialize()
	//a.Run()
}
