package types

import "github.com/ethereum/go-ethereum/core/types"

// pub struct Block {
//     pub header: Header,
//     pub body: Vec<TransactionSigned>,
//     pub ommers: Vec<Header>,
//     pub withdrawals: Option<Withdrawals>,
//     pub requests: Option<Requests>,
// }

type Block struct {
	Header      *Header              `json:"header" gencodec:"required"`
	Body        []*TransactionSigned `json:"body" gencodec:"required"`
	Ommers      []*Header            `json:"ommers" gencodec:"required"`
	Withdrawals *types.Withdrawals   `json:"withdrawals"`
	Requests    *[]*Request          `json:"requests"`
}

func (b *Block) Origin() *types.Block {
	// TODO: implement
	panic("not implemented")
}
