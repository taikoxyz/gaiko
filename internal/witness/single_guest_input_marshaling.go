package witness

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	gaikoTypes "github.com/taikoxyz/gaiko/internal/types"
	"github.com/taikoxyz/gaiko/pkg/mpt"
)

type singleGuestInputJSON struct {
	Block           *gaikoTypes.Block                `json:"block"`
	ChainSpec       *ChainSpec                       `json:"chain_spec"`
	ParentHeader    *gaikoTypes.Header               `json:"parent_header"`
	ParentStateTrie *mpt.MptNode                     `json:"parent_state_trie"`
	ParentStorage   map[common.Address]*StorageEntry `json:"parent_storage"`
	Contracts       []hexutil.Bytes                  `json:"contracts"`
	AncestorHeaders []*gaikoTypes.Header             `json:"ancestor_headers"`
	Taiko           *taikoGuestInputJSON             `json:"taiko"`
}

func (g *singleGuestInputJSON) GethType() *SingleGuestInput {
	if g == nil {
		log.Warn("missing guestInputJSON when converting to GethType")
		return nil
	}
	contracts := make([][]byte, len(g.Contracts))
	for i, contract := range g.Contracts {
		contracts[i] = contract
	}

	ancestorHeaders := make([]*types.Header, len(g.AncestorHeaders))
	for i, ancestorHeader := range g.AncestorHeaders {
		ancestorHeaders[i] = ancestorHeader.GethType()
	}
	return &SingleGuestInput{
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
	BlockProposed  *blockProposedJSON            `json:"block_proposed"`
	ProverData     *TaikoProverData              `json:"prover_data"`
	BlobCommitment *[commitmentSize]byte         `json:"blob_commitment"`
	BlobProof      *[proofSize]byte              `json:"blob_proof"`
	BlobProofType  BlobProofType                 `json:"blob_proof_type"`
}

func (t *taikoGuestInputJSON) GethType() *TaikoGuestInput {
	if t == nil {
		log.Warn("missing taikoGuestInputJSON when converting to GethType")
		return nil
	}
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

type blockProposedJSON struct {
	inner any
}

func (b *blockProposedJSON) GethType() BlockProposed {
	if b == nil {
		log.Warn("missing blockProposedForkJSON when converting to GethType")
		return nil
	}
	switch inner := b.inner.(type) {
	case *gaikoTypes.BlockProposed:
		return NewHeklaBlockProposed(inner.GethType())
	case *gaikoTypes.BlockProposedV2:
		return NewOntakeBlockProposed(inner.GethType())
	case *gaikoTypes.BatchProposed:
		return NewPacayaBlockProposed(inner.GethType())
	case *shastaEventDataJSON:
		return NewShastaBlockProposed(inner.GethType())
	case *NotingBlockProposed:
		return &NotingBlockProposed{}
	default:
		return nil
	}
}

func (b *blockProposedJSON) UnmarshalJSON(data []byte) error {
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
		case "Shasta":
			var inner shastaEventDataJSON
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
	var raw [2]json.RawMessage
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

func (g *SingleGuestInput) UnmarshalJSON(data []byte) error {
	var dec singleGuestInputJSON
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	*g = *dec.GethType()
	return nil
}

type shastaBlobSliceJSON struct {
	BlobHashes []common.Hash `json:"blobHashes"`
	Offset     uint32        `json:"offset"`
	Timestamp  uint64        `json:"timestamp"`
}

type shastaDerivationSourceJSON struct {
	IsForcedInclusion bool                `json:"isForcedInclusion"`
	BlobSlice         shastaBlobSliceJSON `json:"blobSlice"`
}

type shastaDerivationJSON struct {
	OriginBlockNumber  uint64                       `json:"originBlockNumber"`
	OriginBlockHash    common.Hash                  `json:"originBlockHash"`
	BasefeeSharingPctg uint8                        `json:"basefeeSharingPctg"`
	Sources            []shastaDerivationSourceJSON `json:"sources"`
}

type shastaProposalJSON struct {
	ID                             uint64         `json:"id"`
	Timestamp                      uint64         `json:"timestamp"`
	EndOfSubmissionWindowTimestamp uint64         `json:"endOfSubmissionWindowTimestamp"`
	Proposer                       common.Address `json:"proposer"`
	CoreStateHash                  common.Hash    `json:"coreStateHash"`
	DerivationHash                 common.Hash    `json:"derivationHash"`
}

type shastaCoreStateJSON struct {
	NextProposalID              uint64      `json:"nextProposalId"`
	LastProposalBlockID         uint64      `json:"lastProposalBlockId"`
	LastFinalizedProposalID     uint64      `json:"lastFinalizedProposalId"`
	LastCheckpointTimestamp     uint64      `json:"lastCheckpointTimestamp"`
	LastFinalizedTransitionHash common.Hash `json:"lastFinalizedTransitionHash"`
	BondInstructionsHash        common.Hash `json:"bondInstructionsHash"`
}

type shastaEventDataJSON struct {
	Proposal   shastaProposalJSON   `json:"proposal"`
	Derivation shastaDerivationJSON `json:"derivation"`
	CoreState  shastaCoreStateJSON  `json:"core_state"`
	Proposer   common.Address       `json:"proposer"`
}

func (s *shastaEventDataJSON) GethType() *ShastaEventData {
	sources := make([]ShastaDerivationSource, 0, len(s.Derivation.Sources))
	for _, source := range s.Derivation.Sources {
		sources = append(sources, ShastaDerivationSource{
			IsForcedInclusion: source.IsForcedInclusion,
			BlobSlice: ShastaBlobSlice{
				BlobHashes: source.BlobSlice.BlobHashes,
				Offset:     source.BlobSlice.Offset,
				Timestamp:  source.BlobSlice.Timestamp,
			},
		})
	}
	return &ShastaEventData{
		Proposal: ShastaProposal{
			ID:                             s.Proposal.ID,
			Timestamp:                      s.Proposal.Timestamp,
			EndOfSubmissionWindowTimestamp: s.Proposal.EndOfSubmissionWindowTimestamp,
			Proposer:                       s.Proposal.Proposer,
			CoreStateHash:                  s.Proposal.CoreStateHash,
			DerivationHash:                 s.Proposal.DerivationHash,
		},
		Derivation: ShastaDerivation{
			OriginBlockNumber:  s.Derivation.OriginBlockNumber,
			OriginBlockHash:    s.Derivation.OriginBlockHash,
			BasefeeSharingPctg: s.Derivation.BasefeeSharingPctg,
			Sources:            sources,
		},
		CoreState: ShastaCoreState{
			NextProposalID:              s.CoreState.NextProposalID,
			LastProposalBlockID:         s.CoreState.LastProposalBlockID,
			LastFinalizedProposalID:     s.CoreState.LastFinalizedProposalID,
			LastCheckpointTimestamp:     s.CoreState.LastCheckpointTimestamp,
			LastFinalizedTransitionHash: s.CoreState.LastFinalizedTransitionHash,
			BondInstructionsHash:        s.CoreState.BondInstructionsHash,
		},
		Proposer: s.Proposer,
	}
}
