package types

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

type Block struct {
	Header      *Header               `json:"header"      gencodec:"required"`
	Body        TransactionSignedList `json:"body"        gencodec:"required"`
	Ommers      Headers               `json:"ommers"      gencodec:"required"`
	Withdrawals types.Withdrawals     `json:"withdrawals"`
}

func (b *Block) GethType() *types.Block {
	return types.NewBlock(b.Header.GethType(), &types.Body{
		Transactions: b.Body.GethType(),
		Uncles:       b.Ommers.GethType(),
		Withdrawals:  b.Withdrawals,
	}, nil, trie.NewStackTrie(nil))
}
