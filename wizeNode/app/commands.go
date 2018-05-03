package app

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli"

	"wizeBlock/wizeNode/blockchain"
	"wizeBlock/wizeNode/crypto"
	"wizeBlock/wizeNode/wallet"
	//"bitbucket.org/udt/wizefs/internal/command"
	//"bitbucket.org/udt/wizefs/internal/tlog"
)

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug output",
	},
	cli.StringFlag{
		Name:   "nodeADD",
		Value:  "localhost",
		Usage:  "",
		EnvVar: "NODE_ADD",
	},
	cli.StringFlag{
		Name:   "nodeID",
		Value:  "3000",
		Usage:  "",
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
				Value: "3000",
				Usage: "",
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
				Value: "3000",
				Usage: "",
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
				Value: "4000",
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
		//tlog.Debug.Enabled = true
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
		log.Panic("ERROR: Sender address is not valid")
	}
	if !crypto.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
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
		SendTx(KnownNodes[0], nodeID, tx)
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
	nodeADD := c.GlobalString("nodeADD")
	nodeID := c.GlobalString("nodeID")
	fmt.Printf("Starting node %s:%s\n", nodeADD, nodeID)

	minerAddress := c.String("miner")
	apiAddress := c.String("api")

	if len(minerAddress) > 0 {
		if crypto.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
			//StartServer(nodeID, minerAddress, apiAddress)
			node := NewNode(nodeID)
			node.apiAddr = apiAddress
			node.nodeID = nodeID
			node.nodeADD = nodeADD
			node.Run(minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	//StartServer(nodeID, minerAddress, apiAddress)

	node := NewNode(nodeID)
	node.apiAddr = apiAddress
	node.nodeID = nodeID
	node.nodeADD = nodeADD
	node.Run(minerAddress)
	return nil
}
