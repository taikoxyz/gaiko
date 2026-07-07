package witness

import "github.com/ethereum/go-ethereum/common"

type RawProof struct {
	Proof []byte      `json:"proof"`
	Input common.Hash `json:"input"`
}

type RawAggregationGuestInput struct {
	Proofs []*RawProof `json:"proofs"`
}

type ShastaRawAggregationGuestInput struct {
	Proofs            []*RawProof      `json:"proofs"`
	ProofCarryDataVec []ProofCarryData `json:"proof_carry_data_vec"`
}
