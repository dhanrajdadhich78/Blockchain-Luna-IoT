package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"wizeBlock/wizeNode/crypto"
	"wizeBlock/wizeNode/wallet"
)

var blockApi = NewBlockApi()

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
	cli.StringFlag{
		Name:   "nodeID",
		Value:  "3000",
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
	{
		Name:    "getwallet",
		Aliases: []string{"gwal"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "address",
				Usage: "Wallet address",
			},
		},
		Usage:  "Get Wallet ADDRESS info",
		Action: CmdGetWalletInfo,
	},
	// blockchain commands
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
	//	{
	//		Name:    "printchain",
	//		Aliases: []string{"print"},
	//		Usage:   "Print all the blocks of the blockchain",
	//		Action:  CmdPrintChain,
	//	},
	//	{
	//		Name:    "getblock",
	//		Aliases: []string{"gblk"},
	//		Flags: []cli.Flag{
	//			cli.StringFlag{
	//				Name:  "hash",
	//				Usage: "",
	//			},
	//		},
	//		Usage:  "Get a block with hash",
	//		Action: CmdGetBlock,
	//	},
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
	wallets, _ := wallet.NewWalletsExt("wallet%s.dat", nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	walletNew := wallets.GetWallet(address)

	fmt.Println("Your address:", address)
	fmt.Println("Private key: ", hex.EncodeToString(walletNew.GetPrivateKey()))
	fmt.Println("Public key:  ", hex.EncodeToString(walletNew.GetPublicKey()))
	return nil
}

func CmdListAddresses(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	var addresses []string = []string{}
	wallets, err := wallet.NewWalletsExt("wallet%s.dat", nodeID)
	if err != nil {
		return err
	}
	addresses = wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
	return nil
}

func CmdGetBalance(c *cli.Context) (err error) {
	address := c.String("address")
	walletInfo, err := blockApi.GetWalletInfo(address)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	fmt.Printf("Wallet info: %+v\n", walletInfo)
	return nil
}

func CmdGetWalletInfo(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	address := c.String("address")
	wallets, _ := wallet.NewWalletsExt("wallet%s.dat", nodeID)
	walletInfo := wallets.GetWallet(address)
	fmt.Println("Your address:", address)
	fmt.Println("Private key: ", hex.EncodeToString(walletInfo.GetPrivateKey()))
	fmt.Println("Public key:  ", hex.EncodeToString(walletInfo.GetPublicKey()))
	return nil
}

// blockchain commands
func CmdSend(c *cli.Context) (err error) {
	nodeID := c.GlobalString("nodeID")
	from := c.String("from")
	to := c.String("to")
	amount := c.Int("amount")
	mineNow := c.Bool("mine")

	if !crypto.ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		return fmt.Errorf("ERROR: Sender address is not valid")
	}
	if !crypto.ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
		return fmt.Errorf("ERROR: Recipient address is not valid")
	}

	wallets, _ := wallet.NewWalletsExt("wallet%s.dat", nodeID)
	walletInfo := wallets.GetWallet(from)
	pubKey := hex.EncodeToString(walletInfo.GetPublicKey())
	privKeyStruct, err := crypto.GetPrivateKey(nil, walletInfo.GetPrivateKey())
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	prepare := &PrepareTxRequest{
		From:   from,
		To:     to,
		Amount: amount,
		PubKey: pubKey,
	}
	prepared, err := blockApi.PostTxPrepare(prepare)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	// sign
	signatures := make([]string, 0)
	for _, hash := range prepared.Hashes {
		hashToSign, _ := hex.DecodeString(hash)
		r, s, err := crypto.Sign(rand.Reader, privKeyStruct, hashToSign[:])
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return err
		}
		signature := append(r.Bytes(), s.Bytes()...)
		signatures = append(signatures, hex.EncodeToString(signature))
	}

	fmt.Printf("Signatures: %+v\n", signatures)

	sign := &SignTxRequest{
		Txid:       prepared.Txid,
		Minenow:    mineNow,
		Signatures: signatures,
	}
	signed, err := blockApi.PostTxSign(sign)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	if signed.Success {

	}
	return nil
}

// blockchain explorer commands
func CmdPrintChain(c *cli.Context) (err error) {
	return nil
}

func CmdGetBlock(c *cli.Context) (err error) {
	return nil
}
