package witness

import (
	"encoding/json"
	"iter"
	"math/big"
	"slices"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/gaiko/internal/keccak"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

var _ GuestDriver = (*BatchGuestInput)(nil)
var _ json.Unmarshaler = (*BatchGuestInput)(nil)

type BatchGuestInput struct {
	Inputs []*GuestInput
	Taiko  *TaikoGuestBatchInput
}

type TaikoGuestBatchInput struct {
	BatchID            uint64
	L1Header           *types.Header
	BatchProposed      BlockProposedFork
	ChainSpec          *ChainSpec
	ProverData         *TaikoProverData
	TxDataFromCalldata []byte
	TxDataFromBlob     []*[eth.BlobSize]byte
	BlobCommitments    *[][commitmentSize]byte
	BlobProofs         *[][proofSize]byte
	BlobProofType      BlobProofType
}

func (g *BatchGuestInput) GuestInputs() iter.Seq[*Pair] {
	return func(yield func(*Pair) bool) {
		batchProposed := g.Taiko.BatchProposed
		blobDataBufs := g.Taiko.TxDataFromBlob
		var compressedTxListBuf []byte
		for _, blobDataBuf := range blobDataBufs {
			blob := (*eth.Blob)(blobDataBuf)
			data, err := blob.ToData()
			if err != nil {
				log.Warn("Parse blob data failed", "err", err)
				return
			}
			compressedTxListBuf = append(compressedTxListBuf, data...)
		}
		offset, length := batchProposed.BlobTxSliceParam()
		chainID := big.NewInt(int64(g.ChainID()))
		firstBlock := g.Inputs[0].Block.Number()
		txListBytes, err := sliceTxList(firstBlock, compressedTxListBuf, offset, length)
		if err != nil {
			log.Warn(
				"Invalid txlist offset and size in metadata",
				"blockID", firstBlock.Uint64(),
				"err", err,
			)
			return
		}
		txs := decompressTxList(txListBytes, true, true, chainID)
		blockParams := batchProposed.BlockParams()
		next := 0
		for i, blockParam := range blockParams {
			numTransactions := int(blockParam.NumTransactions)
			_txs := []*types.Transaction{g.Inputs[i].Taiko.AnchorTx}
			_txs = append(_txs, txs[next:next+numTransactions]...)
			if !yield(&Pair{g.Inputs[i], _txs}) {
				return
			}
			next += numTransactions
		}
	}
}

func (g *BatchGuestInput) BlockProposedFork() BlockProposedFork {
	return g.Taiko.BatchProposed
}

func (g *BatchGuestInput) verifyBatchModeBlobUsage(proofType ProofType) error {
	blobProofType := getBlobProofType(proofType, g.Taiko.BlobProofType)
	for i := range len(g.Taiko.TxDataFromBlob) {
		blob := g.Taiko.TxDataFromBlob[i]
		commitment := (*g.Taiko.BlobCommitments)[i]
		proof := (*g.Taiko.BlobProofs)[i]
		if err := verifyBlob(blobProofType, (*eth.Blob)(blob), commitment, (*kzg4844.Proof)(&proof)); err != nil {
			return err
		}
	}
	return nil
}

func (g *BatchGuestInput) calculatePacayaTxsHash(
	txListHash common.Hash,
	blobHashes [][32]byte,
) (common.Hash, error) {
	data, err := batchTxHashArgs.Pack(txListHash, blobHashes)
	if err != nil {
		return common.Hash{}, err
	}
	return keccak.Keccak(data), nil
}

func (g *BatchGuestInput) BlockMetadataFork(proofType ProofType) (BlockMetadataFork, error) {
	if err := g.verifyBatchModeBlobUsage(proofType); err != nil {
		return nil, err
	}
	for input := range slices.Values(g.Inputs) {
		if err := defaultSupportedChainSpecs.verifyChainSpec(input.ChainSpec); err != nil {
			return nil, err
		}

	}
	txListHash := keccak.Keccak(g.Taiko.TxDataFromCalldata)
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
		BaseFeeConfig:      *g.Taiko.BatchProposed.BaseFeeConfig(),
		BlobCreatedIn:      g.Taiko.BatchProposed.BlobCreatedIn(),
	}

	data, err := batchInfoComponentsArgs.Pack(batchInfo)
	if err != nil {
		return nil, err
	}
	infoHash := keccak.Keccak(data)

	return NewPacayaBlockMetadata(&pacaya.ITaikoInboxBatchMetadata{
		InfoHash:   infoHash,
		Proposer:   g.Taiko.BatchProposed.Proposer(),
		BatchId:    g.Taiko.BatchID,
		ProposedAt: g.Taiko.BatchProposed.ProposedAt(),
	}), nil

}

func (g *BatchGuestInput) Transition() any {
	return &pacaya.ITaikoInboxTransition{
		ParentHash: g.Inputs[0].ParentHeader.Hash(),
		BlockHash:  g.Inputs[len(g.Inputs)-1].Block.Hash(),
		StateRoot:  g.Inputs[len(g.Inputs)-1].ParentHeader.Root,
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
