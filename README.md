# Install. Setup. Start

**go version go1.9.2+**

**linux/mac os x required (_Doesn't support windows now_)**

For start cluster:
- install docker-ce
- run bash start_cluster.sh

connect to blockchain node on localhost:4000 // postman collection can be found in root folder of project

connect to raft node on localhost:11001  //postman collection comming soon

# How It Works

## Blockchain Basic: blocks, transactions


Blockchain is just a database with certain structure: it’s an ordered, back-linked list. Which means that blocks are stored in the insertion order and that each block is linked to the previous one. This structure allows to quickly get the latest block in a chain and to (efficiently) get a block by its hash.

In blockchain it’s blocks that store valuable information, in particular, transactions, the essence of any cryptocurrency. Besides this, a block contains some technical information, like its version, current timestamp and the hash of the previous block.

A transaction is a combination of inputs and outputs. Inputs of a new transaction reference outputs of a previous transaction. Outputs are where coins are actually stored.


## Blockchain in Details: wallets, merkle tree, utxo


In WizeBlock, your identity is a pair of private and public keys stored on your computer (or stored in some other place you have access to). A wallet is nothing but a key pair. In the construction of wallet a new key pair is generated with ECDSA which is based on elliptic curves. A private key is generated using the curve, and a public key is generated from the private key. One thing to notice: in elliptic curve based algorithms, public keys are points on a curve. Thus, a public key is a combination of X, Y coordinates.

If you want to send coins to someone, you need to know their address. But addresses (despite being unique) are not something that identifies you as the owner of a “wallet”. In fact, such addresses are a human readable representation of public keys. The address generation algorithm utilizes a combination of open algorithms that takes a public key and returns real Base58-based address.

Currently WizeBlock has generating wallets on the WizeBlock node side, but in the next version wallets will generate on the user side in the Desktop application.


### merklee tree
### utxo


## Blockchain TODO



# Network

## Network nodes and their roles


Current WizeBlock network implementation has some simplification. Bitcoint network is decentralized, which means there’re no servers that do stuff and clients that use servers to get or process data. In our implementation, there will be centralization though.

We’ll have three node roles:
- The central node. This is the node all other nodes will connect to, and this is the node that’ll sends data between other nodes.
- A miner node. This node will store new transactions in mempool and when there’re enough of transactions, it’ll mine a new block.
- A wallet node. This node will be used to send coins between wallets. It’ll store a full copy of blockchain.


## tcp service


Nodes communicate by the means of messages.

When a new node is run, it gets several nodes from a DNS seed, and sends them **version** message. If  node is not the central one, it must send version message to the central node to find out if its blockchain is outdated.

Next message **getblocks** means “show me what blocks you have” (in Bitcoin, it’s more complex). Pay attention, it doesn’t say “give me all your blocks”, instead it requests a list of block hashes.

WizeBlock uses **inv** to show other nodes what blocks or transactions current node has. Again, it doesn’t contain whole blocks and transactions, just their hashes.

Message **getdata** is a request for certain block or transaction, and it can contain only one block or transaction ID. The handler is straightforward: if they request a block, return the block; if they request a transaction, return the transaction. Notice, that we don’t check if we actually have this block or transaction.

Messages **block** and **tx** actually transfer the data.


## rest service





## network todo

# Mining

## mining algorithm - proof of work
## mining todo

# Encryption

## hashing algoritms and hashcash - 1-2*
## wallet key pair generating - 5*
## transaction signing and verification - 5-6*
## encryption todo

# Furure Features

