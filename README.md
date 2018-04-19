# Install. Setup. Start

**go version go1.9.2+**

**linux/mac os x required (_Doesn't support windows now_)**

For start cluster:
- install Docker-CE: https://www.docker.com/community-edition
- install Docker Compose: https://docs.docker.com/compose/install/
- run bash start_cluster.sh

Connect URL to blockchain node is localhost:4000.
Connect URL to raft node is localhost:11001.

Postman collection can be found in root folder of project.


# How It Works


## Blockchain Basic: blocks, transactions


Blockchain is just a database with certain structure: it’s an ordered, back-linked list. Which means that blocks are stored in the insertion order and that each block is linked to the previous one. This structure allows to quickly get the latest block in a chain and to (efficiently) get a block by its hash.

In blockchain it’s blocks that store valuable information, in particular, transactions, the essence of any cryptocurrency. Besides this, a block contains some technical information, like its version, current timestamp and the hash of the previous block.

A transaction is a combination of inputs and outputs. Inputs of a new transaction reference outputs of a previous transaction. Outputs are where coins are actually stored.


## Blockchain in Details: wallets


In WizeBlock, your identity is a pair of private and public keys stored on your computer (or stored in some other place you have access to). A wallet is nothing but a key pair. In the construction of wallet a new key pair is generated with ECDSA which is based on elliptic curves. A private key is generated using the curve, and a public key is generated from the private key. One thing to notice: in elliptic curve based algorithms, public keys are points on a curve. Thus, a public key is a combination of X, Y coordinates.

If you want to send coins to someone, you need to know their address. But addresses (despite being unique) are not something that identifies you as the owner of a “wallet”. In fact, such addresses are a human readable representation of public keys. The address generation algorithm utilizes a combination of open algorithms that takes a public key and returns real Base58-based address.

Currently WizeBlock has generating wallets on the WizeBlock node side, but in the next version wallets will generate on the user side in the Desktop application.


# Network


## Network nodes and their roles


Current WizeBlock network implementation has some simplification. Bitcoint network is decentralized, which means there’re no servers that do stuff and clients that use servers to get or process data. In our implementation, there will be centralization though.

We’ll have three node roles:
- The central node. This is the node all other nodes will connect to, and this is the node that’ll sends data between other nodes.
- A miner node. This node will store new transactions in mempool and when there’re enough of transactions, it’ll mine a new block.
- A wallet node. This node will be used to send coins between wallets. It’ll store a full copy of blockchain.


## Network Messages


Nodes communicate by the means of messages.

When a new node is run, it gets several nodes from a DNS seed, and sends them **version** message. If  node is not the central one, it must send version message to the central node to find out if its blockchain is outdated.

Next message **getblocks** means “show me what blocks you have” (in Bitcoin, it’s more complex). Pay attention, it doesn’t say “give me all your blocks”, instead it requests a list of block hashes.

WizeBlock uses **inv** to show other nodes what blocks or transactions current node has. Again, it doesn’t contain whole blocks and transactions, just their hashes.

Message **getdata** is a request for certain block or transaction, and it can contain only one block or transaction ID. The handler is straightforward: if they request a block, return the block; if they request a transaction, return the transaction. Notice, that we don’t check if we actually have this block or transaction.

Messages **block** and **tx** actually transfer the data.


## REST Service


WizeBlock provides a REST service with next API:
- Create Wallet (nodeAddress:nodePort/wallet/new) returns wallet info (private and public keys, base58-based address)
- Get Wallet (nodeAddress:nodePort/wallet/{wallet_address}) returns wallet details (wallet balance)
- Send Transaction (nodeAddress:nodePort/send) with POST parameters: from_address, to_address, amount value and minenow flag; minenow flag is used for mining new blocks, if it is true new block will mine, and if it false the Miner nodes receives the transaction and keeps it in its memory pool and when there are enough transactions in the memory pool, the miner starts mining a new block


## Network todo


# Mining


## Mining algorithm, proof-of-work


A key idea of blockchain is that one has to perform some hard work to put data in it. It is this hard work that makes blockchain secure and consistent. Also, a reward is paid for this hard work (this is how people get coins for mining).

In blockchain, some participants (miners) of the network work to sustain the network, to add new blocks to it, and get a reward for their work. As a result of their work, a block is incorporated into the blockchain in a secure way, which maintains the stability of the whole blockchain database. It’s worth noting that, the one who finished the work has to prove this.

This whole “do hard work and prove” mechanism is called proof-of-work. It’s hard because it requires a lot of computational power: even high performance computers cannot do it quickly.

Proof-of-Work algorithms must meet a requirement: doing the work is hard, but verifying the proof is easy. A proof is usually handed to someone else, so for them, it shouldn’t take much time to verify it.

WizeBlock uses Hashcash, a Proof-of-Work algorithm that was initially developed to prevent email spam. This is a brute force algorithm: you change the counter, calculate a new hash, check it, increment the counter, calculate a hash, etc. That’s why it’s computationally expensive.

More about Hashcash: https://en.wikipedia.org/wiki/Hashcash


## Mining todo


# Encryption


## Hashing algoritms


Hashing is a process of obtaining a hash for specified data. A hash is a unique representation of the data it was calculated on. A hash function is a function that takes data of arbitrary size and produces a fixed size hash. Hashing functions are widely used to check the consistency of data. Some software providers publish checksums in addition to a software package.

In blockchain, hashing is used to guarantee the consistency of a block. The input data for a hashing algorithm contains the hash of the previous block, thus making it impossible (or, at least, quite difficult) to modify a block in the chain: one has to recalculate its hash and hashes of all the blocks after it.

More about hashing: https://en.bitcoin.it/wiki/Block_hashing_algorithm


## ECDSA


WizeBlock uses elliptic curves to generate private keys. Elliptic curves is a complex mathematical concept, which we’re not going to explain in details here. What we need to know is that these curves can be used to generate really big and random numbers. The curve used by WizeBlock can randomly pick a number between 0 and 2²⁵⁶ (which is approximately 10⁷⁷, when there are between 10⁷⁸ and 10⁸² atoms in the visible universe). Such a huge upper limit means that it’s almost impossible to generate the same private key twice.

Also, WizeBlock uses ECDSA (Elliptic Curve Digital Signature Algorithm) algorithm to sign transactions.


# Future Features

