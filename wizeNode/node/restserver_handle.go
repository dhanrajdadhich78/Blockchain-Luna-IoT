package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/crypto"
	"wizeBlock/wizeNode/core/wallet"
)

// TODO: refactoring - names, funcs
// TODO: add logging & error handling
// TODO: actual and deprecated APIs

type Prepare struct {
	From   string
	To     string
	Amount int
	PubKey string
}

type Sign struct {
	//From       string
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

func (s *RestServer) sayHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize "+s.node.NodeAddress.String())
}

func (s *RestServer) getWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	resp := map[string]interface{}{
		"success": true,
		"credit":  s.node.blockchain.GetWalletBalance(hash),
		//"credit":  GetWalletCredits(hash, node.nodeID, node.blockchain),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// DEPRECATED: inner usage
func (s *RestServer) deprecatedWalletsList(w http.ResponseWriter, r *http.Request) {
	wallets, err := wallet.NewWallets(s.node.NodeID)
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
func (s *RestServer) deprecatedWalletCreate(w http.ResponseWriter, r *http.Request) {
	wallets, _ := wallet.NewWallets(s.node.NodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(s.node.NodeID)
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
func (s *RestServer) deprecatedSend(w http.ResponseWriter, r *http.Request) {
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

	if from == to {
		fmt.Println("ERROR: Sender address is equal to Recipient address")
		return
	}
	if !crypto.ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		return
	}
	if !crypto.ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
		return
	}

	UTXOSet := blockchain.UTXOSet{s.node.blockchain}

	wallets, err := wallet.NewWallets(s.node.NodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	if wallet == nil {
		fmt.Println("The Address doesn't belongs to you!")
		return
	}

	tx := blockchain.NewUTXOTransaction(wallet, to, amount, &UTXOSet)

	respsuccess := true

	//
	currentNodeAddress := s.node.NodeAddress
	fmt.Printf("currentNodeAddress: %s\n", currentNodeAddress)

	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := s.node.blockchain.MineBlock(txs)
		if newBlock != nil {
			UTXOSet.Update(newBlock)
		} else {
			respsuccess = false
		}
	} else {
		// TODO: проверять остаток на балансе с учетом незамайненых транзакций,
		// во избежание двойного использования выходов
		knownNode0 := s.node.Network.Nodes[0]
		s.node.Client.SendTx(knownNode0, tx)
	}

	//
	if respsuccess {
		for _, node := range s.node.Network.Nodes {
			fmt.Printf("node: %s\n", node)
			if !node.CompareToAddress(currentNodeAddress) {
				s.node.Client.SendVersion(node, s.node.blockchain.GetBestHeight())
			}
		}
	}

	resp := map[string]interface{}{
		"success": respsuccess,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// send transaction steps: prepare/sign
func (s *RestServer) prepare(w http.ResponseWriter, r *http.Request) {
	var prepare Prepare
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Failed to read the request body: %v\n", err)
		sendErrorMessage(w, "Failed to read the request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &prepare); err != nil {
		fmt.Printf("Could not decode the request body as JSON: %v\n", err)
		sendErrorMessage(w, "Could not decode the request body as JSON", http.StatusBadRequest)
		return
	}

	from := prepare.From
	to := prepare.To
	amount := prepare.Amount
	pubKey, _ := hex.DecodeString(prepare.PubKey)

	fmt.Printf("from: %s, to: %s, amount: %d\n", from, to, amount)
	fmt.Printf("pubkey: %s, pubkeyHex: %x\n", prepare.PubKey, pubKey)

	if from == "" || to == "" || amount <= 0 {
		sendErrorMessage(w, "Please check your prepare request", http.StatusBadRequest)
		return
	}

	if from == to {
		fmt.Println("ERROR: Sender address is equal to Recipient address")
		sendErrorMessage(w, "Sender address is equal to Recipient address", http.StatusBadRequest)
		return
	}

	if !crypto.ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		sendErrorMessage(w, "Sender address is not valid", http.StatusBadRequest)
		return
	}
	if !crypto.ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
		sendErrorMessage(w, "Recipient address is not valid", http.StatusBadRequest)
		return
	}

	UTXOSet := blockchain.UTXOSet{s.node.blockchain}

	tx, txToSign, err := blockchain.PrepareUTXOTransaction(from, to, amount, pubKey, &UTXOSet)
	if err != nil || tx == nil || txToSign == nil {
		sendErrorMessage(w, "Could not prepare transaction", http.StatusInternalServerError)
		return
	}

	//txid := fmt.Sprintf("%x", txToSign.TxID)

	fmt.Printf("tx.ID: %x\n", tx.ID)
	fmt.Printf("txToSign.ID: %x\n", txToSign.TxID)

	txid := hex.EncodeToString(txToSign.TxID)

	// add to Prepared-Transactions
	preparedTx := &PreparedTransaction{
		From:        from,
		Transaction: tx,
	}
	s.node.preparedTxs[txid] = preparedTx

	fmt.Printf("txid: %s, hashesToSign count: %d\n", txid, len(txToSign.HashesToSign))

	resp := map[string]interface{}{
		"success": true,
		"txid":    txid,
		"hashes":  txToSign.HashesToSign,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (s *RestServer) sign(w http.ResponseWriter, r *http.Request) {
	var sign Sign
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Failed to read the request body: %v\n", err)
		sendErrorMessage(w, "Failed to read the request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &sign); err != nil {
		fmt.Printf("Could not decode the request body as JSON: %v\n", err)
		sendErrorMessage(w, "Could not decode the request body as JSON", http.StatusBadRequest)
		return
	}

	//from := sign.From
	txid := sign.TxID
	signatures := sign.Signatures
	mineNow := sign.MineNow

	fmt.Printf("txid: %s, signatures count: %d\n", txid, len(signatures))

	if txid == "" {
		sendErrorMessage(w, "Please check your sign request", http.StatusBadRequest)
		return
	}

	UTXOSet := blockchain.UTXOSet{s.node.blockchain}
	TxID, _ := hex.DecodeString(txid)

	// get from Prepared Transactions
	preparedTx, ok := s.node.preparedTxs[txid]
	from := preparedTx.From

	fmt.Printf("TxID: %x\n", TxID)
	fmt.Printf("preparedTx: %x\n", preparedTx.Transaction.ID)
	fmt.Printf("preparedTx From: %s\n", from)

	if !ok {
		fmt.Println("Could not get transaction by txid")
		sendErrorMessage(w, "Could not get transaction by txid", http.StatusBadRequest)
		return
	}
	fmt.Println("GOOD: Get transaction by txid!")

	// check from
	if !crypto.ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		sendErrorMessage(w, "Sender address is not valid", http.StatusBadRequest)
		return
	}

	txSignatures := &blockchain.TransactionWithSignatures{
		TxID:       TxID,
		Signatures: signatures,
	}

	respsuccess := true

	tx := blockchain.SignUTXOTransaction(preparedTx.Transaction, txSignatures, &UTXOSet)

	// network update
	currentNodeAddress := s.node.NodeAddress
	fmt.Printf("currentNodeAddress: %s\n", currentNodeAddress)

	// mining block: now and with miner's help
	if mineNow {
		// TODO: minenow=true
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := s.node.blockchain.MineBlock(txs)
		if newBlock != nil {
			UTXOSet.Update(newBlock)
		} else {
			respsuccess = false
		}

		// network update
		if respsuccess {
			for _, node := range s.node.Network.Nodes {
				fmt.Printf("node: %s\n", node)
				if !node.CompareToAddress(currentNodeAddress) {
					s.node.Client.SendVersion(node, s.node.blockchain.GetBestHeight())
				}
			}
		}
	} else {
		// TODO: minenow=false

		// TODO: проверять остаток на балансе с учетом незамайненых транзакций,
		// во избежание двойного использования выходов

		knownNode0 := s.node.Network.Nodes[0]
		fmt.Printf("Send Tx: %x, from %s, to: %s\n", tx.ID, knownNode0, currentNodeAddress)
		s.node.Client.SendTx(knownNode0, tx)
	}

	// remove from Prepared-Transactions
	delete(s.node.preparedTxs, txid)

	resp := map[string]interface{}{
		"success": respsuccess,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

// inner usage
func (s *RestServer) printBlockchain(w http.ResponseWriter, r *http.Request) {

	bci := s.node.blockchain.Iterator()
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

func (s *RestServer) getBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockHash := vars["hash"]
	//TODO: зачем итеретор? попробовать выбрать по ключу
	bci := s.node.blockchain.Iterator()
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
	response, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("ERROR: respondWithJSON: %v\n", err)
	}

	fmt.Printf("Payload: %v\n", payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)

	//fmt.Printf("response: %v\n", response)
}

type appError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	HttpStatus int    `json:"status"`
	ExitCode   int    `json:"exitcode"`
}

type errorResource struct {
	Data appError `json:"data"`
}

func displayAppError(w http.ResponseWriter, handlerError error, message string, code int, exitCode int) {
	errObj := appError{
		Error:      "nil",
		Message:    message,
		HttpStatus: code,
		ExitCode:   exitCode,
	}

	if handlerError != nil {
		errObj.Error = handlerError.Error()
	}

	fmt.Printf("[app error]: %s\n", handlerError)

	respondWithJSON(w, code, errorResource{Data: errObj})
}

func sendErrorMessage(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(status)
	io.WriteString(w, msg)
}

//type ErrorResponse struct {
//	Error string `json:"error"`
//}

//func (node *Node) writeResponse(w http.ResponseWriter, b []byte) {
//	w.Header().Set("Content-Type", "application/json")
//	w.Write(b)
//}

//func (node *Node) error(w http.ResponseWriter, err error, message string) {
//	node.logError(err)

//	b, err := json.Marshal(&ErrorResponse{
//		Error: message,
//	})
//	if err != nil {
//		node.logError(err)
//	}

//	node.writeResponse(w, b)
//}
