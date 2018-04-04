package blockchain

import (
	"os"
	"strconv"
	"testing"

	"github.com/boltdb/bolt"
)

const testNodeID = "2999"

func clearData() {
	err := os.Remove("files/db/wizebit_" + testNodeID + ".db")
	if err != nil {
		return
	}
	err = os.Remove("files/wallets/wallet_" + testNodeID + ".dat")
	if err != nil {
		return
	}

	os.RemoveAll("files")
}

func createWallet() string {
	wallets, _ := NewWallets(testNodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(testNodeID)
	return address
}

func createBlockchain(t *testing.T, address string) {
	if !ValidateAddress(address) {
		t.Fatal("ERROR: Address is not valid")
	}
	bc := CreateBlockchain(address, testNodeID)
	defer bc.Db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	t.Log("Blockchain creating: Done!")
}

func printChain(t *testing.T, bc *Blockchain) {
	//bc := NewBlockchain(testNodeID)
	//defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		t.Logf("============ Block %x ============\n", block.Hash)
		t.Logf("Height: %d\n", block.Height)
		t.Logf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
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

func sendTransaction(t *testing.T, bc *Blockchain, from, to string, amount int) {
	if !ValidateAddress(from) {
		t.Fatal("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		t.Fatal("ERROR: Recipient address is not valid")
	}

	UTXOSet := UTXOSet{bc}

	wallets, err := NewWallets(testNodeID)
	if err != nil {
		t.Fatal(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(wallet, to, amount, &UTXOSet)

	mineNow := true
	if mineNow {
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
		//} else {
		//	sendTx(knownNodes[0], tx)
	}

	t.Log("Success!")
}

func newBlock(t *testing.T, bc *Blockchain) *Block {
	var lastHash []byte
	var lastHeight int

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return NewBlock(nil, lastHash, lastHeight+1)
}

func testBlockchainJustCreating(t *testing.T) {
	clearData()

	os.MkdirAll("files/db", 0775)
	os.MkdirAll("files/wallets", 0775)

	// createwallet?
	walletAddress := createWallet()
	t.Logf("Wallet address: %s\n", walletAddress)

	// createblockchain -> Blockchain.CreateBlockchain
	createBlockchain(t, walletAddress)

	// Blockchain.NewBlockchain
	bc := NewBlockchain(testNodeID)
	defer bc.Db.Close()
	t.Logf("Blockchain TIP: %x\n", bc.tip)

	// printchain -> Iterator
	printChain(t, bc)

	t.Log("Passed")
	clearData()
}

func testBlockchainCreatingAndAddingBlock(t *testing.T) {
	clearData()

	os.MkdirAll("files/db", 0775)
	os.MkdirAll("files/wallets", 0775)

	// wallets
	walletAddress1 := createWallet()
	t.Logf("Wallet1 address: %s\n", walletAddress1)
	walletAddress2 := createWallet()
	t.Logf("Wallet2 address: %s\n", walletAddress2)

	// createblockchain -> Blockchain.CreateBlockchain
	createBlockchain(t, walletAddress1)

	// Blockchain.NewBlockchain
	bc := NewBlockchain(testNodeID)
	defer bc.Db.Close()

	t.Logf("Blockchain TIP: %x\n", bc.tip)

	// Blockchain.AddBlock + Block.NewBlock
	sendTransaction(t, bc, walletAddress1, walletAddress2, 1)
	//block := newBlock(t, bc)
	//bc.AddBlock(block)

	// Blockchain.GetBlock
	// Blockchain.GetBlockHashes
	// Blockchain.GetBestHeight?

	// printchain -> Iterator
	printChain(t, bc)

	t.Log("Passed")
	clearData()
}
