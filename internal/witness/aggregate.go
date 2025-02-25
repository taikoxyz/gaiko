package witness

import "github.com/ethereum/go-ethereum/common"

type RawProof struct {
	Proof []byte      `json:"proof"`
	Input common.Hash `json:"input"`
}

type RawAggregationGuestInput struct {
	Proofs []*RawProof `json:"proofs"`
}
