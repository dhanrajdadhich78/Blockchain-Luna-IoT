package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/crypto"
	"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
	"wizeBlock/wizeNode/core/wallet"
	"wizeBlock/wizeNode/node"
)

// TODO: all commands should moved to another file module
//       & perhaps it will be Node struct module
// TODO: [DC] check all base commands
//       createwallet, createblockchain, getbalance, send
//       listaddresses, printchain, getblock, startnode
// TODO: [DC] add all node commands
// TODO: [DC] add all blockchain commands
// TODO: optimize all commands, add command groups, subcommands?

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug output",
	},
	cli.StringFlag{
		Name:   "nodeADD",
		Value:  "localhost",
		Usage:  "Node address",
		EnvVar: "NODE_ADD",
	},
	cli.IntFlag{
		Name:   "nodeID",
		Value:  3000,
		Usage:  "Node ID (port)",
		EnvVar: "NODE_ID",
	},
}

var Commands = []cli.Command{
	// wallet commands
	{
		Name:    "createwallet",
		Aliases: []string{"cw"},
		Usage:   "Generates a new key-pair and saves it into the wallet file",
		Action:  CmdCreateWallet,
	},
	{
		Name:    "listaddresses",
		Aliases: []string{"la"},
		Usage:   "Lists all addresses from the wallet file",
		Action:  CmdListAddresses,
	},
	{
		Name:    "getbalance",
		Aliases: []string{"gbal"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "address",
				Usage: "Wallet address",
			},
		},
		Usage:  "Get balance of ADDRESS",
		Action: CmdGetBalance,
	},
	// blockchain commands
	{
		Name:    "createblockchain",
		Aliases: []string{"cbc"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "address",
				Usage: "Wallet address for miner rewards",
			},
		},
		Usage:  "Create a blockchain and send genesis block reward to ADDRESS",
		Action: CmdCreateBlockchain,
	},
	{
		Name:    "send",
		Aliases: []string{"send"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from",
				Usage: "Sender ADDRESS",
			},
			cli.StringFlag{
				Name:  "to",
				Usage: "Recepient ADDRESS",
			},
			cli.IntFlag{
				Name:  "amount",
				Usage: "Amount of coins",
			},
			cli.BoolFlag{
				Name:  "mine",
				Usage: "Mine in the same node or only with miner nodes",
			},
		},
		Usage:  "Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set",
		Action: CmdSend,
	},
	// blockchain explorer commands
	{
		Name:    "printchain",
		Aliases: []string{"print"},
		Usage:   "Print all the blocks of the blockchain",
		Action:  CmdPrintChain,
	},
	{
		Name:    "getblock",
		Aliases: []string{"gblk"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "hash",
				Usage: "",
			},
		},
		Usage:  "Get a block with hash",
		Action: CmdGetBlock,
	},
	// p2p network commands
	{
		Name:    "startnode",
		Aliases: []string{"start"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "miner",
				Usage: "",
			},
			cli.StringFlag{
				Name:  "api",
				Value: ":4000",
				Usage: "",
			},
			cli.IntFlag{
				Name:  "pause",
				Value: 0,
				Usage: "",
			},
		},
		Usage:  "Start a node with ID specified in NODE_ID env. var. -miner enables mining",
		Action: CmdStartNode,
	},
}

// CommandNotFound implements action when subcommand not found
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.\n", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

// CommandBefore implements action before run command
func CommandBefore(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.Debug.Enabled = true
	}
	return nil
}

// wallet commands
func CmdCreateWallet(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	wallets, _ := wallet.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	walletNew := wallets.GetWallet(address)

	fmt.Printf("Your new address: %s\n", address)
	fmt.Println("Private key: ", hex.EncodeToString(walletNew.GetPrivateKey()))
	fmt.Println("Public key: ", hex.EncodeToString(walletNew.GetPublicKey()))
	return nil
}

func CmdListAddresses(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	var addresses []string = []string{}
	if nodeID == "3100" {
		// FIXME: experiment with get addresses
		bc := blockchain.NewBlockchain("3000")
		addresses = bc.GetAddresses()
	} else {
		wallets, err := wallet.NewWallets(nodeID)
		if err != nil {
			return err
		}
		addresses = wallets.GetAddresses()
	}
	for _, address := range addresses {
		fmt.Println(address)
	}
	return nil
}

func CmdGetBalance(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	address := c.String("address")
	bc := blockchain.NewBlockchain(nodeID)
	balance := bc.GetWalletBalance(address)
	fmt.Printf("Balance of '%s': %d\n", address, balance)
	return nil
}

// blockchain commands
func CmdCreateBlockchain(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	address := c.String("address")
	if !crypto.ValidateAddress(address) {
		fmt.Printf("ERROR: Address is not valid")
		return fmt.Errorf("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
	return nil
}

func CmdSend(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	from := c.String("from")
	to := c.String("to")
	amount := c.Int("amount")
	mineNow := c.Bool("mine")

	if !crypto.ValidateAddress(from) {
		log.Fatal.Println("ERROR: Sender address is not valid")
		return
	}
	if !crypto.ValidateAddress(to) {
		log.Fatal.Println("ERROR: Recipient address is not valid")
		return
	}

	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Fatal.Printf("Error: %s", err)
		return
	}
	wallet := wallets.GetWallet(from)

	if wallet == nil {
		fmt.Println("The Address doesn't belongs to you!")
		return
	}

	tx := blockchain.NewUTXOTransaction(wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		// TODO: проверять остаток на балансе с учетом незамайненых транзакций,
		// во избежание двойного использования выходов
		//SendTx(KnownNodes[0], nodeID, tx)
	}

	fmt.Println("Success!")
	return nil
}

// blockchain explorer commands
func CmdPrintChain(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		fmt.Printf("Created at: %s\n", time.Unix(block.Timestamp, 0))
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil
}

func CmdGetBlock(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Db.Close()

	blockHash := c.String("hash")
	bci := bc.Iterator()

	for {
		block := bci.Next()

		hash := fmt.Sprintf("%x", block.Hash)
		if hash == blockHash {
			fmt.Printf("============ Block %x ============\n", block.Hash)
			fmt.Printf("Height: %d\n", block.Height)
			fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
			fmt.Printf("Created at : %s\n", time.Unix(block.Timestamp, 0))
			pow := blockchain.NewProofOfWork(block)
			fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
			for _, tx := range block.Transactions {
				fmt.Println(tx)
			}
			fmt.Printf("\n\n")
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil
}

// p2p network commands
func CmdStartNode(c *cli.Context) (err error) {
	pause := c.Int("pause")
	if pause > 0 {
		time.Sleep(time.Duration(pause) * time.Second)
	}

	nodeID := c.GlobalInt("nodeID")
	nodeIDStr := strconv.Itoa(nodeID)
	nodeAddr := network.NodeAddr{
		Host: c.GlobalString("nodeADD"),
		Port: nodeID,
	}
	log.Info.Printf("Starting Node %s", nodeAddr)

	// PROD: add request to masternode and get nodeID
	//nodeAddress := os.Getenv("NODE_ADD") + ":" + nodeIDStr

	// PROD: add request to masternode and get nodeID
	//nodeAddress := os.Getenv("NODE_ADD") + ":" + nodeIDStr

	// FIXME: it is just apiPort
	apiAddr := c.String("api")

	// FIXME: minerWalletAddress to Node, not to NodeServer
	minerWalletAddress := c.String("miner")

	// register server in masternode
	registerDigest()

	if len(minerWalletAddress) > 0 {
		if crypto.ValidateAddress(minerWalletAddress) {
			log.Info.Println("Mining is on. Address to receive rewards: ", minerWalletAddress)
		} else {
			log.Warn.Println("Wrong miner address!")
			return fmt.Errorf("Wrong miner address!")
		}
	}

	newNode := node.NewNode(nodeIDStr, nodeAddr, apiAddr, minerWalletAddress)
	newNode.Run()
	return nil
}
