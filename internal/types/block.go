package types

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type Block struct {
	Header      *Header               `json:"header"      gencodec:"required"`
	Body        TransactionSignedList `json:"body"        gencodec:"required"`
	Ommers      Headers               `json:"ommers"      gencodec:"required"`
	Withdrawals types.Withdrawals     `json:"withdrawals"`
}

func (b *Block) GethType() *types.Block {
	if b == nil {
		log.Warn("missing Block when converting to GethType")
		return nil
	}
	return types.NewBlockWithHeader(b.Header.GethType()).WithBody(types.Body{
		Transactions: b.Body.GethType(),
		Uncles:       b.Ommers.GethType(),
		Withdrawals:  b.Withdrawals,
	})
}
