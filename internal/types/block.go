package types

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

type Block struct {
	Header      *Header            `json:"header"      gencodec:"required"`
	Body        TransactionSigneds `json:"body"        gencodec:"required"`
	Ommers      Headers            `json:"ommers"      gencodec:"required"`
	Withdrawals types.Withdrawals  `json:"withdrawals"`
	Requests    Requests           `json:"requests"`
}

func (b *Block) GethType() *types.Block {
	return types.NewBlock(b.Header.GethType(), &types.Body{
		Transactions: b.Body.GethType(),
		Uncles:       b.Ommers.GethType(),
		Withdrawals:  b.Withdrawals,
		Requests:     b.Requests.GethType(),
	}, nil, trie.NewStackTrie(nil))
}
