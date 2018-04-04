package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	blockchain "wizeBlock/wizeNode/blockchain"
)

type Prepare struct {
	From   string
	To     string
	Amount int
	PubKey string
}

type Sign struct {
	From       string
	TxID       string
	Signatures []string
	MineNow    bool
}

// DEPRECATED: inner usage
type Send struct {
	From    string
	To      string
	Amount  int
	MineNow bool
}

func (node *Node) sayHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize "+node.nodeADD)
}

func (node *Node) getWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	resp := map[string]interface{}{
		"success": true,
		"credit":  GetWalletCredits(hash, node.nodeID, node.blockchain),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// DEPRECATED: inner usage
func (node *Node) listWallets(w http.ResponseWriter, r *http.Request) {
	wallets, err := blockchain.NewWallets(node.nodeID)
	if err != nil {
		log.Panic(err)
	}

	resp := map[string]interface{}{
		"success":     true,
		"listWallets": wallets.GetAddresses(),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// DEPRECATED: inner usage
func (node *Node) createWallet(w http.ResponseWriter, r *http.Request) {
	wallets, _ := blockchain.NewWallets(node.nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(node.nodeID)
	wallet := wallets.GetWallet(address)

	//fmt.Printf("Your new address: %s\n", address)
	//fmt.Println("Private key: ", hex.EncodeToString(wallet.GetPrivateKey()))
	//fmt.Println("Public key: ", hex.EncodeToString(wallet.GetPublicKey()))

	resp := map[string]interface{}{
		"success": true,
		"address": address,
		"privkey": hex.EncodeToString(wallet.GetPrivateKey()),
		"pubkey":  hex.EncodeToString(wallet.GetPublicKey()),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// DEPRECATED: inner usage
func (node *Node) send(w http.ResponseWriter, r *http.Request) {
	//func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {

	var send Send
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read the request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &send); err != nil {
		sendErrorMessage(w, "Could not decode the request body as JSON", http.StatusBadRequest)
		return
	}
	from := send.From
	to := send.To
	amount := send.Amount
	mineNow := send.MineNow
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	UTXOSet := blockchain.UTXOSet{node.blockchain}

	wallets, err := blockchain.NewWallets(node.nodeID)
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

		newBlock := node.blockchain.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		// TODO: проверять остаток на балансе с учетом незамайненых транзакций,
		// во избежание двойного использования выходов
		SendTx(KnownNodes[0], node.nodeID, tx)
	}

	resp := map[string]interface{}{
		"success": true,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// send transaction steps: prepare/sign
func (node *Node) prepare(w http.ResponseWriter, r *http.Request) {
	var prepare Prepare
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Failed to read the request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &prepare); err != nil {
		fmt.Printf("Could not decode the request body as JSON: %v", err)
		sendErrorMessage(w, "Could not decode the request body as JSON", http.StatusBadRequest)
		return
	}

	from := prepare.From
	to := prepare.To
	amount := prepare.Amount
	pubKey, _ := hex.DecodeString(prepare.PubKey)

	fmt.Printf("from: %s, to: %s, amount: %d\n", from, to, amount)
	fmt.Printf("pubkey: %s, pubkeyHex: %x\n", prepare.PubKey, pubKey)

	if !blockchain.ValidateAddress(from) {
		//log.Panic("ERROR: Sender address is not valid")
		fmt.Println("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		//log.Panic("ERROR: Recipient address is not valid")
		fmt.Println("ERROR: Recipient address is not valid")
	}

	UTXOSet := blockchain.UTXOSet{node.blockchain}

	tx, txToSign := blockchain.PrepareUTXOTransaction(from, to, amount, pubKey, &UTXOSet)

	//txid := fmt.Sprintf("%x", txToSign.TxID)
	txid := hex.EncodeToString(txToSign.TxID)

	// add to Prepared-Transactions
	node.preparedTxs[txid] = tx

	fmt.Printf("txid: %s, dataToSign count: %d\n", txid, len(txToSign.DataToSign))

	resp := map[string]interface{}{
		"success": true,
		"txid":    txid,
		"data":    txToSign.DataToSign,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) sign(w http.ResponseWriter, r *http.Request) {
	var sign Sign
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Failed to read the request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &sign); err != nil {
		fmt.Printf("Could not decode the request body as JSON: %v", err)
		sendErrorMessage(w, "Could not decode the request body as JSON", http.StatusBadRequest)
		return
	}

	from := sign.From
	txid := sign.TxID
	signatures := sign.Signatures
	mineNow := sign.MineNow

	fmt.Printf("from: %s, txid: %s, signatures count: %d\n", from, txid, len(signatures))

	UTXOSet := blockchain.UTXOSet{node.blockchain}

	TxID, _ := hex.DecodeString(txid)

	// get from Prepared-Transactions
	preparedTx, ok := node.preparedTxs[txid]
	if !ok {
		fmt.Println("Could not get transaction by txid")
		sendErrorMessage(w, "Could not get transaction by txid", http.StatusBadRequest)
		return
	}
	fmt.Println("GOOD: Get transaction by txid!")

	txSignatures := &blockchain.TransactionWithSignatures{
		TxID:       TxID,
		Signatures: signatures,
	}

	tx := blockchain.SignUTXOTransaction(preparedTx, txSignatures, &UTXOSet)
	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := node.blockchain.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		// TODO: проверять остаток на балансе с учетом незамайненых транзакций,
		// во избежание двойного использования выходов
		SendTx(KnownNodes[0], node.nodeID, tx)
	}

	// remove from Prepared-Transactions
	delete(node.preparedTxs, txid)

	resp := map[string]interface{}{
		"success": true,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// inner usage
func (node *Node) printBlockchain(w http.ResponseWriter, r *http.Request) {

	bci := node.blockchain.Iterator()
	chain := make([]*blockchain.Block, 0)

	for {
		block := bci.Next()

		//fmt.Printf("============ Block %x ============\n", block.Hash)
		//fmt.Printf("Height: %d\n", block.Height)
		//fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		//fmt.Printf("Created at: %s\n", time.Unix(block.Timestamp, 0))
		//pow := b.NewProofOfWork(block)
		//fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		//for _, tx := range block.Transactions {
		//	fmt.Println(tx)
		//}
		//fmt.Printf("\n\n")
		chain = append(chain, block)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	resp := map[string]interface{}{
		"success":   true,
		"chainlist": chain,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) getBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockHash := vars["hash"]
	//TODO: зачем итеретор? попробовать выбрать по ключу
	bci := node.blockchain.Iterator()
	var result *blockchain.Block

	for {
		block := bci.Next()

		bh, _ := json.Marshal(block.Hash)
		hash := string(bh[1 : len(bh)-1])

		if hash == blockHash {
			//fmt.Printf("============ Block %x ============\n", block.Hash)
			//fmt.Printf("Height: %d\n", block.Height)
			//fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
			//fmt.Printf("Created at : %s\n", time.Unix(block.Timestamp, 0))
			//pow := b.NewProofOfWork(block)
			//fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
			//for _, tx := range block.Transactions {
			//	fmt.Println(tx)
			//}
			//fmt.Printf("\n\n")
			result = block
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	resp := map[string]interface{}{
		"success": true,
		"credit":  result,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func sendErrorMessage(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(status)
	io.WriteString(w, msg)
}
