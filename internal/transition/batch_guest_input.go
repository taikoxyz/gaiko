package transition

import (
	"encoding/json"
	"iter"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

var _ GuestDriver = (*BatchGuestInput)(nil)
var _ json.Unmarshaler = (*BatchGuestInput)(nil)

type BatchGuestInput struct {
	Inputs []*GuestInput
	Taiko  *TaikoGuestBatchInput
}

type TaikoGuestBatchInput struct {
	BatchId            uint64
	L1Header           *types.Header
	BatchProposed      BlockProposedFork
	ChainSpec          *ChainSpec
	ProverData         *TaikoProverData
	TxDataFromCalldata []byte
	TxDataFromBlob     [][eth.BlobSize]byte
	BlobCommitments    *[][commitmentSize]byte
	BlobProofs         *[][proofSize]byte
	BlobProofType      BlobProofType
}

func (g *BatchGuestInput) UnmarshalJSON(data []byte) error {
	// TODO: Implement
	return json.Unmarshal(data, g)
}

func (g *BatchGuestInput) GuestInputs() iter.Seq[Pair] {
	panic("not implemented") // TODO: Implement
}

func (g *BatchGuestInput) BlockProposedFork() BlockProposedFork {
	return g.Taiko.BatchProposed
}

func (g *BatchGuestInput) verifyBatchModeBlobUsage(proofType ProofType) error {
	blobProofType := getBlobProofType(proofType, g.Taiko.BlobProofType)
	for i := 0; i < len(g.Taiko.TxDataFromBlob); i++ {
		blob := g.Taiko.TxDataFromBlob[i]
		commitment := (*g.Taiko.BlobCommitments)[i]
		proof := (*g.Taiko.BlobProofs)[i]
		if err := verifyBlob(blobProofType, blob, commitment, (*kzg4844.Proof)(&proof)); err != nil {
			return err
		}
	}
	return nil
}

func (g *BatchGuestInput) calculatePacayaTxsHash(txListHash common.Hash, blobHashes [][32]byte) (common.Hash, error) {
	data, err := batchTxHashArgs.Pack(txListHash, blobHashes)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(keccak(data)), nil
}

func (g *BatchGuestInput) BlockMetaDataFork(proofType ProofType) (BlockMetaDataFork, error) {
	if err := g.verifyBatchModeBlobUsage(proofType); err != nil {
		return nil, err
	}
	txListHash := common.BytesToHash(keccak(g.Taiko.TxDataFromCalldata))
	txsHash, err := g.calculatePacayaTxsHash(txListHash, g.Taiko.BatchProposed.BlobHashes())
	if err != nil {
		return nil, err
	}

	blocks := make([]pacaya.ITaikoInboxBlockParams, len(g.Inputs))

	for i, input := range g.Inputs {
		signalSlots, err := decodeAnchorV3Args_signalSlots(input.Taiko.AnchorTx.Data()[4:])
		if err != nil {
			return nil, err
		}
		blockParams := pacaya.ITaikoInboxBlockParams{
			NumTransactions: uint16(input.Block.Transactions().Len()) - 1,
			TimeShift:       uint8(input.Block.Time() - g.Taiko.BatchProposed.ProposedAt()),
			SignalSlots:     signalSlots,
		}
		blocks[i] = blockParams
	}

	batchInfo := &pacaya.ITaikoInboxBatchInfo{
		TxsHash:            txsHash,
		Blocks:             blocks,
		BlobHashes:         g.Taiko.BatchProposed.BlobHashes(),
		ExtraData:          g.Taiko.BatchProposed.ExtraData(),
		Coinbase:           g.Taiko.BatchProposed.Coinbase(),
		ProposedIn:         g.Taiko.BatchProposed.ProposedIn(),
		BlobByteOffset:     g.Taiko.BatchProposed.BlobTxListOffset(),
		BlobByteSize:       g.Taiko.BatchProposed.BlobTxListLength(),
		GasLimit:           g.Taiko.BatchProposed.GasLimit(),
		LastBlockId:        g.Inputs[len(g.Inputs)-1].Block.NumberU64(),
		LastBlockTimestamp: g.Inputs[len(g.Inputs)-1].Block.Time(),
		AnchorBlockId:      g.Taiko.L1Header.Number.Uint64(),
		AnchorBlockHash:    g.Taiko.L1Header.Hash(),
		BaseFeeConfig:      g.Taiko.BatchProposed.BaseFeeConfig(),
	}

	data, err := batchInfoComponentsArgs.Pack(batchInfo)
	if err != nil {
		return nil, err
	}
	infoHash := keccak(data)

	return NewPacayaBlockMetadata(&pacaya.ITaikoInboxBatchMetadata{
		InfoHash:   common.BytesToHash(infoHash),
		Proposer:   g.Taiko.BatchProposed.Proposer(),
		BatchId:    g.Taiko.BatchId,
		ProposedAt: g.Taiko.BatchProposed.ProposedAt(),
	}), nil

}

func (g *BatchGuestInput) Transition() *ontake.TaikoDataTransition {
	return &ontake.TaikoDataTransition{
		ParentHash: g.Inputs[0].ParentHeader.Hash(),
		BlockHash:  g.Inputs[len(g.Inputs)-1].Block.Hash(),
		StateRoot:  g.Inputs[len(g.Inputs)-1].ParentHeader.Root,
		Graffiti:   g.Taiko.ProverData.Graffiti,
	}
}

func (g *BatchGuestInput) ForkVerifierAddress(proofType ProofType) (common.Address, error) {
	return g.Taiko.ChainSpec.getForkVerifierAddress(g.Taiko.BatchProposed.BlockNumber(), proofType)
}

func (g *BatchGuestInput) Prover() common.Address {
	return g.Taiko.ProverData.Prover
}

func (g *BatchGuestInput) ChainID() uint64 {
	return g.Taiko.ChainSpec.ChainID
}

func (g *BatchGuestInput) IsTaiko() bool {
	return g.Taiko.ChainSpec.IsTaiko
}

func (g *BatchGuestInput) ChainConfig() (*params.ChainConfig, error) {
	return g.Taiko.ChainSpec.chainConfig()
}
