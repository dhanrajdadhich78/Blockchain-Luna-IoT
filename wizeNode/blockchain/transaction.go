package blockchain

import (
	"bytes"
	//"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	ecdsa "wizeBlock/wizeNode/crypto"
	"wizeBlock/wizeNode/utils"
)

const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TransactionToSign struct {
	TxID       []byte
	DataToSign []string
	Signatures []string
}

type TransactionWithSignatures struct {
	TxID       []byte
	Signatures []string
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) PrepareToSign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) *TransactionToSign {
	if tx.IsCoinbase() {
		return nil
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	prepareToSign := TransactionToSign{
		TxID:       tx.ID,
		DataToSign: make([]string, len(txCopy.Vin)),
		Signatures: make([]string, len(txCopy.Vin)),
	}

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		//// Signing
		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		fmt.Printf("A signature: %s\n", hex.EncodeToString(signature))

		prepareToSign.DataToSign[inID] = dataToSign
		prepareToSign.Signatures[inID] = hex.EncodeToString(signature)

		// another function
		//tx.Vin[inID].Signature = signature
		//txCopy.Vin[inID].PubKey = nil
	}

	return &prepareToSign
}

func (tx *Transaction) SignPrepared(txSignatures *TransactionWithSignatures, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, _ := range txCopy.Vin {
		//prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		//txCopy.Vin[inID].Signature = nil
		//txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		signature, _ := hex.DecodeString(txSignatures.Signatures[inID])
		fmt.Printf("B signature: %x\n", signature)

		tx.Vin[inID].Signature = signature
		//txCopy.Vin[inID].PubKey = nil
	}
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {
		pubKeyHash := HashPubKey(input.PubKey)
		versionedPayload := append([]byte{Version}, pubKeyHash...)
		fullPayload := append(versionedPayload, Checksum(versionedPayload)...)

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
		lines = append(lines, fmt.Sprintf("       Addr  :    %s", utils.Base58Encode(fullPayload)))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
		lines = append(lines, fmt.Sprintf("       Addr  : %s", output.Address))
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		//outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash, vout.Address})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			return false, fmt.Errorf("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false, fmt.Errorf("ERROR: Verify return false")
		}
		txCopy.Vin[inID].PubKey = nil
	}

	return true, nil
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// NewEmissionCoinbaseTX creates a new coinbase transaction with emission
func NewEmissionCoinbaseTX(to, data string, emission int) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(emission, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// PrepareUTXOTransaction prepare a new transaction
func PrepareUTXOTransaction(from, to string, amount int, pubKey []byte, privKey []byte, UTXOSet *UTXOSet) (*Transaction, *TransactionToSign) {
	var inputs []TXInput
	var outputs []TXOutput

	//
	ECDSAKeyX := pubKey[:64]
	ECDSAKeyY := pubKey[64:]

	keyD := new(big.Int)
	keyX := new(big.Int)
	keyY := new(big.Int)
	keyD.SetBytes(privKey)
	keyX.SetBytes(ECDSAKeyX)
	keyY.SetBytes(ECDSAKeyY)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     keyX,
		Y:     keyY,
	}
	privateKey := ecdsa.PrivateKey{D: keyD, PublicKey: publicKey}

	pubKeyHash := HashPubKey(pubKey)
	fmt.Printf("pubKeyHash %x\n", pubKeyHash)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	// OLDTODO: delete
	fmt.Printf("Sum of outputs %d\n", acc)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// TODO: find pubKey by pubKeyHash
	//pubKey := pubKeyHash

	// Build a list of inputs
	// OLDTODO: rewrite to smart choice of outputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			// OLDTODO: delete
			//fmt.Println("Output", out)
			input := TXInput{txID, out, nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	fmt.Printf("tx.ID: %x\n", tx.ID)

	return &tx, UTXOSet.Blockchain.PrepareTransactionToSign(&tx, privateKey)
}

// SignUTXOTransaction signs a prepared transaction
func SignUTXOTransaction(preparedTx *Transaction, txSignatures *TransactionWithSignatures, UTXOSet *UTXOSet) *Transaction {
	UTXOSet.Blockchain.SignPreparedTransaction(preparedTx, txSignatures)
	return preparedTx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	if wallet == nil {
		return nil
	}

	pubKeyHash := HashPubKey(wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	// OLDTODO: delete
	//fmt.Printf("Sum of outputs %d\n", acc)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	// OLDTODO: rewrite to smart choice of outputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			// OLDTODO: delete
			//fmt.Println("Output", out)
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		fmt.Println(err)
	}
	return transaction
}
