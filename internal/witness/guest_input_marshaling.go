package witness

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/taikoxyz/gaiko/internal/mpt"
	gaikoTypes "github.com/taikoxyz/gaiko/internal/types"
)

type guestInputJSON struct {
	Block           *gaikoTypes.Block                `json:"block"`
	ChainSpec       *ChainSpec                       `json:"chain_spec"`
	ParentHeader    *gaikoTypes.Header               `json:"parent_header"`
	ParentStateTrie *mpt.MptNode                     `json:"parent_state_trie"`
	ParentStorage   map[common.Address]*StorageEntry `json:"parent_storage"`
	Contracts       []hexutil.Bytes                  `json:"contracts"`
	AncestorHeaders []*gaikoTypes.Header             `json:"ancestor_headers"`
	Taiko           *taikoGuestInputJSON             `json:"taiko"`
}

func (g *guestInputJSON) GethType() *GuestInput {
	contracts := make([][]byte, len(g.Contracts))
	for i, contract := range g.Contracts {
		contracts[i] = contract
	}

	ancestorHeaders := make([]*types.Header, len(g.AncestorHeaders))
	for i, ancestorHeader := range g.AncestorHeaders {
		ancestorHeaders[i] = ancestorHeader.GethType()
	}
	return &GuestInput{
		Block:           g.Block.GethType(),
		ChainSpec:       g.ChainSpec,
		ParentHeader:    g.ParentHeader.GethType(),
		ParentStateTrie: g.ParentStateTrie,
		ParentStorage:   g.ParentStorage,
		Contracts:       contracts,
		AncestorHeaders: ancestorHeaders,
		Taiko:           g.Taiko.GethType(),
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

func (t *taikoGuestInputJSON) GethType() *TaikoGuestInput {
	return &TaikoGuestInput{
		L1Header:       t.L1Header.GethType(),
		TxData:         t.TxData,
		AnchorTx:       t.AnchorTx.GethType(),
		BlockProposed:  t.BlockProposed.GethType(),
		ProverData:     t.ProverData,
		BlobCommitment: t.BlobCommitment,
		BlobProof:      t.BlobProof,
		BlobProofType:  t.BlobProofType,
	}
}

type blockProposedForkJSON struct {
	inner interface{}
}

func (b *blockProposedForkJSON) GethType() BlockProposedFork {
	switch inner := b.inner.(type) {
	case *gaikoTypes.BlockProposed:
		return NewHeklaBlockProposed(inner.GethType())
	case *gaikoTypes.BlockProposedV2:
		return NewOntakeBlockProposed(inner.GethType())
	case *gaikoTypes.BatchProposed:
		return NewPacayaBlockProposed(inner.GethType())
	case *NotingBlockProposed:
		return &NotingBlockProposed{}
	default:
		return nil
	}
}

func (b *blockProposedForkJSON) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key, val := range raw {
		switch key {
		case "Hekla":
			var inner gaikoTypes.BlockProposed
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			b.inner = &inner
		case "Ontake":
			var inner gaikoTypes.BlockProposedV2
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			b.inner = &inner
		case "Pacaya":
			var inner gaikoTypes.BatchProposed
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			b.inner = &inner
		case "Nothing":
			b.inner = &NotingBlockProposed{}
		default:
			return fmt.Errorf("unknown BlockProposedFork type: %s", key)
		}
	}
	return nil
}

func (s *StorageEntry) UnmarshalJSON(data []byte) error {
	raw := [2]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var trie mpt.MptNode
	if err := json.Unmarshal(raw[0], &trie); err != nil {
		return err
	}
	s.Trie = &trie
	var slots []*math.HexOrDecimal256
	if err := json.Unmarshal(raw[1], &slots); err != nil {
		return err
	}
	s.Slots = make([]*big.Int, len(slots))
	for i, slot := range slots {
		s.Slots[i] = (*big.Int)(slot)
	}
	return nil
}

func (g *GuestInput) UnmarshalJSON(data []byte) error {
	var dec guestInputJSON
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	*g = *dec.GethType()
	return nil
}
