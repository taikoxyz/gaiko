package transition

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

type StorageEntry struct {
	Trie  trie.Trie
	Slots []*big.Int
}

type GuestInput struct {
	Block           types.Block                     `json:"block"`
	ChainSpec       ChainSpec                       `json:"chain_spec"`
	ParentHeader    types.Header                    `json:"parent_header"`
	ParentStateTrie trie.Trie                       `json:"parent_state_trie"`
	ParentStorage   map[common.Address]StorageEntry `json:"parent_storage"`
	Contracts       [][]byte                        `json:"contracts"`
	AncestorHeaders []types.Header                  `json:"ancestor_headers"`
	Taiko           TaikoGuestInput                 `json:"taiko"`
}

type TaikoGuestInput struct {
	L1Header       types.Header          `json:"l1_header"`
	TxData         []byte                `json:"tx_data"`
	AnchorTx       *types.Transaction    `json:"anchor_tx"`
	BlockProposed  BlockProposedFork     `json:"block_proposed"`
	ProverData     TaikoProverData       `json:"prover_data"`
	BlobCommitment *[commitmentSize]byte `json:"blob_commitment"`
	BlobProof      *[proofSize]byte      `json:"blob_proof"`
	BlobProofType  BlobProofType         `json:"blob_proof_type"`
}

type BlobProofType string

const (
	KzgVersionedHash   BlobProofType = "kzg_versioned_hash"
	ProofOfEquivalence BlobProofType = "proof_of_equivalence"
)

type TaikoProverData struct {
	Prover   common.Address
	Graffiti common.Hash
}
