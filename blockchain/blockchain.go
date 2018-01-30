package blockchain

import (
	"sync"
	"time"
	b "wizeBlockchain/block"
)

type Blockchain struct {
	Blocks b.Blocks
	mu     sync.RWMutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: b.Blocks{b.GenesisBlock},
		mu:     sync.RWMutex{},
	}
}

func (bc *Blockchain) len() int {
	return len(bc.Blocks)
}

func (bc *Blockchain) getGenesisBlock() *b.Block {
	return bc.getBlock(0)
}

func (bc *Blockchain) GetLatestBlock() *b.Block {
	return bc.getBlock(bc.len() - 1)
}

func (bc *Blockchain) getBlock(index int) *b.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.Blocks[index]
}

func (bc *Blockchain) GenerateBlock(data string) *b.Block {
	block := &b.Block{
		Index:        bc.GetLatestBlock().Index + 1,
		PreviousHash: bc.GetLatestBlock().Hash,
		Timestamp:    time.Now().Unix(),
		Data:         data,
	}
	block.Hash = block.Hhash()

	return block
}

func (bc *Blockchain) AddBlock(block *b.Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.Blocks = append(bc.Blocks, block)
}

func (bc *Blockchain) ReplaceBlocks(blocks b.Blocks) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.Blocks = blocks
}

func (bc *Blockchain) isValidGenesisBlock() bool {
	block := bc.getGenesisBlock()

	return block.Hash == b.GenesisBlock.Hash &&
		block.IsValidHash()
}

func (bc *Blockchain) IsValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if bc.len() == 0 {
		return false
	}
	if !bc.isValidGenesisBlock() {
		return false
	}

	prevBlock := bc.getGenesisBlock()
	for i := 1; i < bc.len(); i++ {
		block := bc.getBlock(i)

		if ok := b.IsValidBlock(block, prevBlock); !ok {
			return false
		}

		prevBlock = block
	}

	return true
}