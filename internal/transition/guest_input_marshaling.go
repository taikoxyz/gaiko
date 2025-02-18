package transition

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	gaikoTypes "github.com/taikoxyz/gaiko/internal/types"
)

// pub struct GuestInput {
//     pub block: Block,
//     pub chain_spec: ChainSpec,
//     pub parent_header: Header,
//     pub parent_state_trie: MptNode,
//     pub parent_storage: HashMap<Address, StorageEntry>,
//     pub contracts: Vec<Bytes>,
//     pub ancestor_headers: Vec<Header>,
//     pub taiko: TaikoGuestInput,
// }

type guestInputJSON struct {
	Block           *gaikoTypes.Block               `json:"block"`
	ChainSpec       *ChainSpec                      `json:"chain_spec"`
	ParentHeader    *gaikoTypes.Header              `json:"parent_header"`
	ParentStateTrie *trie.Trie                      `json:"parent_state_trie"`
	ParentStorage   map[common.Address]StorageEntry `json:"parent_storage"`
	Contracts       []hexutil.Bytes                 `json:"contracts"`
	AncestorHeaders []*gaikoTypes.Header            `json:"ancestor_headers"`
	Taiko           *taikoGuestInputJSON            `json:"taiko"`
}

func (g *guestInputJSON) Origin() *GuestInput {
	contracts := make([][]byte, len(g.Contracts))
	for i, contract := range g.Contracts {
		contracts[i] = contract
	}

	ancestorHeaders := make([]*types.Header, len(g.AncestorHeaders))
	for i, ancestorHeader := range g.AncestorHeaders {
		ancestorHeaders[i] = ancestorHeader.Origin()
	}
	return &GuestInput{
		Block:           g.Block.Origin(),
		ChainSpec:       g.ChainSpec,
		ParentHeader:    g.ParentHeader.Origin(),
		ParentStateTrie: g.ParentStateTrie,
		ParentStorage:   g.ParentStorage,
		Contracts:       contracts,
		AncestorHeaders: ancestorHeaders,
		Taiko:           g.Taiko.Origin(),
	}
}

type taikoGuestInputJSON struct {
	L1Header       *gaikoTypes.Header            `json:"l1_header"`
	TxData         []byte                        `json:"tx_data"`
	AnchorTx       *gaikoTypes.TransactionSigned `json:"anchor_tx"`
	BlockProposed  *blockProposedForkJSON        `json:"block_proposed"`
	ProverData     *TaikoProverData              `json:"prover_data"`
	BlobCommitment *[commitmentSize]byte         `json:"blob_commitment"`
	BlobProof      *[proofSize]byte              `json:"blob_proof"`
	BlobProofType  BlobProofType                 `json:"blob_proof_type"`
}

func (t *taikoGuestInputJSON) Origin() *TaikoGuestInput {
	return &TaikoGuestInput{
		L1Header:       t.L1Header.Origin(),
		TxData:         t.TxData,
		AnchorTx:       t.AnchorTx.Origin(),
		BlockProposed:  t.BlockProposed.Origin(),
		ProverData:     t.ProverData,
		BlobCommitment: t.BlobCommitment,
		BlobProof:      t.BlobProof,
		BlobProofType:  t.BlobProofType,
	}
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

type blockProposedForkJSON struct {
	inner interface{}
}

func (b *blockProposedForkJSON) Origin() BlockProposedFork {
	switch inner := b.inner.(type) {
	case *gaikoTypes.BlockProposed:
		return NewHeklaBlockProposed(inner.Origin())
	case *gaikoTypes.BlockProposedV2:
		return NewOntakeBlockProposed(inner.Origin())
	case *gaikoTypes.BatchProposed:
		return NewPacayaBlockProposed(inner.Origin())
	default:
		return nil
	}
}

func (b *blockProposedForkJSON) UnmarshalJSON(input []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(input, &raw); err != nil {
		return err
	}
	for key, val := range raw {
		switch key {
		case "Hekla":
			var bp gaikoTypes.BlockProposed
			if err := json.Unmarshal(val, &bp); err != nil {
				return err
			}
			b.inner = &bp
		case "Ontake":
			var bp gaikoTypes.BlockProposedV2
			if err := json.Unmarshal(val, &bp); err != nil {
				return err
			}
			b.inner = &bp
		case "Pacaya":
			var bp gaikoTypes.BatchProposed
			if err := json.Unmarshal(val, &bp); err != nil {
				return err
			}
			b.inner = &bp
		case "Nothing":
			b.inner = &NotingBlockProposed{}
		default:
			return fmt.Errorf("unknown BlockProposedFork type: %s", key)
		}
	}
	return nil
}

func (g *GuestInput) UnmarshalJSON(data []byte) error {
	var dec guestInputJSON
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	*g = *dec.Origin()
	return nil
}
