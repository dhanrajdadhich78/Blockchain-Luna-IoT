package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"wizeBlock/wizeNode/blockchain"
	"wizeBlock/wizeNode/utils"
)

var (
	input1, input2, input3 chan string
	done1, done2, done3    chan bool
)

func nclearData() {
	err := os.Remove("files/db/wizebit_genesis.db")
	if err != nil {
	}
	clearNodeData("3000")
	clearNodeData("3001")
	clearNodeData("3002")

	os.RemoveAll("files")
}

func clearNodeData(nodeID string) {
	err := os.Remove("files/db/wizebit_" + nodeID + ".db")
	if err != nil {
	}
	err = os.Remove("files/db/wallet_" + nodeID + ".dat")
	if err != nil {
	}
}

func ncreateWallet(nodeID string) string {
	wallets, _ := blockchain.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	return address
}

func ncreateBlockchain(t *testing.T, address, nodeID string) {
	if !blockchain.ValidateAddress(address) {
		t.Fatal("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()

	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	t.Log("Blockchain creating: Done!")
}

func copyFile(src string, dst string) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
	}
}

func nsendTransaction(t *testing.T, bc *blockchain.Blockchain, from, to string, amount int, nodeID string, mineNow bool) {
	if !blockchain.ValidateAddress(from) {
		t.Fatal("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		t.Fatal("ERROR: Recipient address is not valid")
	}

	UTXOSet := blockchain.UTXOSet{bc}

	wallets, err := blockchain.NewWallets(nodeID)
	if err != nil {
		t.Fatal(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewUTXOTransaction(wallet, to, amount, &UTXOSet)

	t.Logf("knownNodes: %v\n", KnownNodes)
	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		SendTx(KnownNodes[0], nodeID, tx)
	}

	t.Log("Success!")
}

func nprintChain(t *testing.T, bc *blockchain.Blockchain) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		t.Logf("============ Block %x ============\n", block.Hash)
		t.Logf("Height: %d\n", block.Height)
		t.Logf("Prev. block: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		t.Logf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			t.Log(tx)
		}
		t.Logf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func ngetBalance(t *testing.T, address, nodeID string) int {
	if !blockchain.ValidateAddress(address) {
		t.Error("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	t.Logf("Balance of '%s': %d\n", address, balance)
	return balance
}

func testScenario1(t *testing.T) {
	nclearData()

	os.MkdirAll("files/db", 0775)
	os.MkdirAll("files/wallets", 0775)

	for index, value := range KnownNodes {
		if value == "wize1:3000" {
			KnownNodes[index] = "localhost:3000"
		}
	}

	input1 := make(chan string)
	defer close(input1)
	input2 := make(chan string)
	defer close(input2)
	input3 := make(chan string)
	defer close(input3)
	done1 := make(chan bool)
	defer close(done1)
	done2 := make(chan bool)
	defer close(done2)
	done3 := make(chan bool)
	defer close(done3)

	var wg sync.WaitGroup
	wg.Add(3)

	var centralAddress, minerAddress string
	var walletAddress1, walletAddress2, walletAddress3 string

	go func() {
		defer wg.Done()
		var nodeID = "3000"
		//var bc *Blockchain
		for {
			value := <-input1
			fmt.Println("node1:", value)

			done1 <- true
			if value == "exit" {
				//bc.db.Close()
				break
			}
			switch value {
			case "step1a":
				// create a wallet and a new blockchain
				centralAddress = ncreateWallet(nodeID)
				t.Logf("centralAddress: %s\n", centralAddress)
				ncreateBlockchain(t, centralAddress, nodeID)

				bc := blockchain.NewBlockchain(nodeID)
				//t.Logf("Blockchain TIP: %x\n", bc.tip)
				//nprintChain(t, bc)
				bc.Db.Close()
			case "step1b":
				copyFile("files/db/wizebit_3000.db", "files/db/wizebit_genesis.db")
			case "step1c": //send some coins from CENTRAL to WALLETS with immediately mining
				bc := blockchain.NewBlockchain(nodeID)
				nsendTransaction(t, bc, centralAddress, walletAddress1, 10, nodeID, true)
				nsendTransaction(t, bc, centralAddress, walletAddress2, 10, nodeID, true)
				//nprintChain(t, bc)
				bc.Db.Close()
			case "step1d":
				// start NODE_ID=3000
			case "step1z":
				// stop NODE_ID=3000
			default:
				t.Log("wrong command")
			}
		}
	}()
	go func() {
		defer wg.Done()
		var nodeID = "3001"
		for {
			value := <-input2
			fmt.Println("node2:", value)

			done2 <- true
			if value == "exit" {
				break
			}
			switch value {
			case "step2a":
				// create three wallets
				walletAddress1 = ncreateWallet(nodeID)
				t.Logf("walletAddress1: %s\n", walletAddress1)
				walletAddress2 = ncreateWallet(nodeID)
				t.Logf("walletAddress2: %s\n", walletAddress2)
				walletAddress3 = ncreateWallet(nodeID)
				t.Logf("walletAddress3: %s\n", walletAddress3)
			case "step2b":
				copyFile("files/db/wizebit_genesis.db", "files/db/wizebit_3001.db")
			case "step2c":
				// start NODE_ID=3001
			case "step2d":
				// stop NODE_ID=3001
			case "step2e":
				// check the balances
				ngetBalance(t, walletAddress1, nodeID)
				ngetBalance(t, walletAddress2, nodeID)
				ngetBalance(t, centralAddress, nodeID)
			case "step2f":
				// send some coins without auto-mining
				bc := blockchain.NewBlockchain(nodeID)
				nsendTransaction(t, bc, walletAddress1, walletAddress3, 3, nodeID, false)
				time.Sleep(5000 * time.Millisecond)
				nsendTransaction(t, bc, walletAddress2, walletAddress3, 6, nodeID, false)
				//nprintChain(t, bc)
				bc.Db.Close()
			case "step2h":
				// check the balances
				ngetBalance(t, walletAddress1, nodeID)
				ngetBalance(t, walletAddress2, nodeID)
				ngetBalance(t, walletAddress3, nodeID)
				ngetBalance(t, centralAddress, nodeID)
			default:
				t.Log("wrong command")
			}
		}
	}()
	go func() {
		defer wg.Done()
		var nodeID = "3002"
		for {
			value := <-input3
			fmt.Println("node3:", value)

			done3 <- true
			if value == "exit" {
				break
			}
			switch value {
			case "step3a":
				minerAddress = ncreateWallet(nodeID)
				t.Logf("minerAddress: %s\n", minerAddress)
				copyFile("files/db/wizebit_genesis.db", "files/db/wizebit_3002.db")
			default:
				t.Log("wrong command")
			}
		}
	}()

	runCommand1 := func(command string) {
		input1 <- command
	}
	runCommand2 := func(command string) {
		input2 <- command
	}
	runCommand3 := func(command string) {
		input3 <- command
	}

	go runCommand1("step1a") // create a wallet and a new blockchain
	<-done1

	go runCommand1("step1b") // copy blockchain as genesis blockchain
	<-done1

	go runCommand2("step2a") // create three wallets
	<-done2
	time.Sleep(2000 * time.Millisecond)

	go runCommand1("step1c") // send some coins from CENTRAL to WALLETS with immediately mining
	<-done1
	time.Sleep(7000 * time.Millisecond)

	//go runCommand1("step1d") // start NODE_ID=3000 - THE NODE MUST BE RUNNING UNTIL THE END OF THE SCENARIO
	//<-done1
	wg.Add(1)
	server1 := NewServerForTest("wize1", "3000", "")
	go func() {
		defer wg.Done()
		t.Log("Starting node 3000")
		server1.Start()
	}()

	//////////

	go runCommand2("step2b") // copy genesis blockchain as blockchain for NODE_ID=3001
	<-done2
	time.Sleep(1000 * time.Millisecond)

	//go runCommand2("step2c") // start NODE_ID=3001 - it will download all the blocks from CENTRAL
	//<-done2
	wg.Add(1)
	server2 := NewServerForTest("wize2", "3001", "")
	go func() {
		defer wg.Done()
		t.Log("Starting node 3001")
		server2.Start()
	}()

	//go runCommand2("step2d") // stop NODE_ID=3001
	//<-done2
	time.Sleep(7000 * time.Millisecond)
	server2.Stop()

	go runCommand2("step2e") // check the balances
	<-done2

	//////////
	fmt.Println("PREPARE 3002")

	go runCommand3("step3a") // prepare NODE_ID=3002 as MINER
	<-done3
	time.Sleep(1000 * time.Millisecond)

	fmt.Println("START 3002")

	//go runCommand3("step3b") // start NODE_ID=3002
	//<-done3
	wg.Add(1)
	server3 := NewServerForTest("wize3", "3002", minerAddress)
	go func() {
		defer wg.Done()
		t.Log("Starting node 3002")
		server3.Start()
	}()
	time.Sleep(1000 * time.Millisecond)

	fmt.Println("SEND SOME COINS WITHOUT AUTO-MINTING")

	go runCommand2("step2f") // send some coins without auto-mining
	<-done2
	time.Sleep(7000 * time.Millisecond)

	fmt.Println("START 3001")

	//go runCommand2("step2g") // start/sleep/stop NODE_ID=3001 - it will ...
	//<-done2
	wg.Add(1)
	server2 = NewServerForTest("wize2", "3001", "")
	go func() {
		defer wg.Done()
		t.Log("Starting node 3001")
		server2.Start()
	}()
	time.Sleep(7000 * time.Millisecond)
	fmt.Println("STOP 3001")
	server2.Stop()

	fmt.Println("CHECK THE BALANCES")
	go runCommand2("step2h") // check the balances
	<-done2

	//////////

	//go runCommand3("step3z") // stop NODE_ID=3002 - THE NODE MUST BE RUNNING UNTIL THE END OF THE SCENARIO
	//<-done3
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("STOP 3002")
	server3.Stop()

	//go runCommand1("step1z") // stop NODE_ID=3000 - THE NODE MUST BE RUNNING UNTIL THE END OF THE SCENARIO
	//<-done1
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("STOP 3000")
	server1.Stop()

	// stop1
	go runCommand1("exit")
	<-done1

	// stop2
	go runCommand2("exit")
	<-done2

	// stop3
	go runCommand3("exit")
	<-done3

	wg.Wait()
	t.Log("finish")
	nclearData()
}
