package blockchain

import (
	"sync"
	"testing"
	b "wizeBlockchain/code_network/block"
)

var testGenesisBlock = &b.Block{
	Index:        0,
	PreviousHash: "0",
	Timestamp:    1465154705,
	Data:         "my genesis block!!",
	Hash:         "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7",
}

func newTestBlockchain(blocks b.Blocks) *Blockchain {
	return &Blockchain{
		Blocks: blocks,
		mu:     sync.RWMutex{},
	}
}

func TestGenerateBlock(t *testing.T) {
	bc := newTestBlockchain(b.Blocks{testGenesisBlock})

	block := bc.generateBlock("white noise")
	if block.Index != bc.GetLatestBlock().Index+1 {
		t.Errorf("want %d but %d", bc.GetLatestBlock().Index+1, block.Index)
	}
	if block.Data != "white noise" {
		t.Errorf("want %q but %q", "white noise", block.Data)
	}
	if block.PreviousHash != bc.GetLatestBlock().Hash {
		t.Errorf("want %q but %q", bc.GetLatestBlock().Hash, block.PreviousHash)
	}
}

func TestAddBlock(t *testing.T) {
	bc := newTestBlockchain(b.Blocks{testGenesisBlock})
	block := &b.Block{
		Index:        1,
		PreviousHash: testGenesisBlock.Hash,
		Timestamp:    1494177351,
		Data:         "white noise",
		Hash:         "1cee23ac6ce3589aedbd92213e0dbf8ab41f8f8e6181a92c1a8243df4b32078b",
	}

	bc.AddBlock(block)
	if bc.len() != 2 {
		t.Fatalf("want %d but %d", 2, bc.len())
	}
	if bc.GetLatestBlock().Hash != block.Hash {
		t.Errorf("want %q but %q", block.Hash, bc.GetLatestBlock().Hash)
	}
}

func TestReplaceBlocks(t *testing.T) {
	bc := newTestBlockchain(b.Blocks{testGenesisBlock})
	blocks := b.Blocks{
		testGenesisBlock,
		&b.Block{
			Index:        1,
			PreviousHash: testGenesisBlock.Hash,
			Timestamp:    1494093545,
			Data:         "white noise",
			Hash:         "1cee23ac6ce3589aedbd92213e0dbf8ab41f8f8e6181a92c1a8243df4b32078b",
		},
	}

	bc.ReplaceBlocks(blocks)
	if bc.len() != 2 {
		t.Fatalf("want %d but %d", 2, bc.len())
	}
	if bc.GetLatestBlock().Hash != blocks[len(blocks)-1].Hash {
		t.Errorf("want %q but %q", blocks[len(blocks)-1].Hash, bc.GetLatestBlock().Hash)
	}
}

type isValidChainTestCase struct {
	name       string
	blockchain *Blockchain
	ok         bool
}

var isValidChainTestCases = []isValidChainTestCase{
	isValidChainTestCase{
		"empty",
		newTestBlockchain(b.Blocks{}),
		false,
	},
	isValidChainTestCase{
		"invalid genesis block",
		newTestBlockchain(b.Blocks{
			&b.Block{
				Index:        0,
				PreviousHash: "0",
				Timestamp:    1465154705,
				Data:         "bad genesis block!!",
				Hash:         "627ab16dbcede0cfa91c85a88c30c4eaae41b8500a961d0d09451323c6e25bf8",
			},
		}),
		false,
	},
	isValidChainTestCase{
		"invalid block",
		newTestBlockchain(b.Blocks{
			testGenesisBlock,
			&b.Block{
				Index:        2,
				PreviousHash: testGenesisBlock.Hash,
				Timestamp:    1494177351,
				Data:         "white noise",
				Hash:         "6e27d73b81b2abf47e6766b8aad12a114614fccac669d0d2162cb842f0484420",
			},
		}),
		false,
	},
	isValidChainTestCase{
		"valid",
		newTestBlockchain(b.Blocks{
			testGenesisBlock,
			&b.Block{
				Index:        1,
				PreviousHash: testGenesisBlock.Hash,
				Timestamp:    1494177351,
				Data:         "white noise",
				Hash:         "1cee23ac6ce3589aedbd92213e0dbf8ab41f8f8e6181a92c1a8243df4b32078b",
			},
		}),
		true,
	},
}

func TestIsValidChain(t *testing.T) {
	for _, testCase := range isValidChainTestCases {
		if ok := testCase.blockchain.IsValid(); ok != testCase.ok {
			t.Errorf("[%s] want %t but %t", testCase.name, testCase.ok, ok)
		}
	}
}