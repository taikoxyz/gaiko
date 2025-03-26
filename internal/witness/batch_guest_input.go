package witness

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"math"
	"math/big"
	"slices"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

var _ WitnessInput = (*BatchGuestInput)(nil)
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
	TxDataFromBlob     [][eth.BlobSize]byte
	BlobCommitments    *[][commitmentSize]byte
	BlobProofs         *[][proofSize]byte
	BlobProofType      BlobProofType
}

func (g *BatchGuestInput) GuestInputs() iter.Seq[*Pair] {
	return func(yield func(*Pair) bool) {
		batchProposed := g.Taiko.BatchProposed
		chainID := big.NewInt(int64(g.ChainID()))
		var txs types.Transactions
		if batchProposed.BlobUsed() {
			var compressedTxListBuf []byte
			for _, blobDataBuf := range g.Taiko.TxDataFromBlob {
				blob := eth.Blob(blobDataBuf)
				if data, err := blob.ToData(); err != nil {
					log.Error("Parse blob data failed", "err", err)
				} else {
					compressedTxListBuf = append(compressedTxListBuf, data...)
				}
			}
			batchID := new(big.Int).SetUint64(g.Taiko.BatchID)
			var err error
			compressedTxListBuf, err = sliceTxList(
				batchID,
				compressedTxListBuf,
				batchProposed.BlobTxSliceParam(),
			)
			if err != nil {
				log.Error(
					"Invalid txlist offset and size in metadata",
					"batchId", batchID,
					"err", err,
				)
			}
			txs = decompressTxList(
				compressedTxListBuf,
				blobMaxTxListBytes,
				batchProposed.BlobUsed(),
				true,
				chainID,
			)
		} else {
			txs = decompressTxList(
				g.Taiko.TxDataFromCalldata,
				calldataMaxTxListBytes,
				batchProposed.BlobUsed(),
				true,
				chainID,
			)
		}

		blockParams := batchProposed.BlockParams()
		next := 0
		for i, blockParam := range blockParams {
			numTxs := int(blockParam.NumTransactions)
			_txs := []*types.Transaction{g.Inputs[i].Taiko.AnchorTx}
			_txs = append(_txs, txs[next:next+numTxs]...)
			if !yield(&Pair{g.Inputs[i], _txs}) {
				return
			}
			next += numTxs
		}
	}
}

func (g *BatchGuestInput) BlockProposedFork() BlockProposedFork {
	return g.Taiko.BatchProposed
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

func (g *BatchGuestInput) Verify(proofType ProofType) error {
	// 1. verify chain spec
	for input := range slices.Values(g.Inputs) {
		if err := defaultSupportedChainSpecs.verifyChainSpec(input.ChainSpec); err != nil {
			return err
		}
	}

	blobProofType := getBlobProofType(proofType, g.Taiko.BlobProofType)
	// 2.1 check the same length of blob's commitments or proofs
	switch blobProofType {
	case KzgVersionedHash:
		if len(g.Taiko.TxDataFromBlob) != 0 &&
			(g.Taiko.BlobCommitments == nil || len(g.Taiko.TxDataFromBlob) != len(*g.Taiko.BlobCommitments)) {
			return fmt.Errorf(
				"invalid blob commitments length, expected: %d, got: %d",
				len(g.Taiko.TxDataFromBlob), len(*g.Taiko.BlobCommitments),
			)
		}
	case ProofOfEquivalence:
		if len(g.Taiko.TxDataFromBlob) != 0 &&
			(g.Taiko.BlobProofs == nil || len(g.Taiko.TxDataFromBlob) != len(*g.Taiko.BlobProofs)) {
			return fmt.Errorf(
				"invalid blob proofs length, expected: %d, got: %d",
				len(g.Taiko.TxDataFromBlob), len(*g.Taiko.BlobCommitments),
			)
		}
	}

	// 2.2 verify the correctness of blob's proofs
	// for i := range len(g.Taiko.TxDataFromBlob) {
	// 	blob := g.Taiko.TxDataFromBlob[i]
	// 	commitment := (*g.Taiko.BlobCommitments)[i]
	// 	proof := (*g.Taiko.BlobProofs)[i]

	// 	if err := verifyBlob(blobProofType, (*eth.Blob)(&blob), commitment, (*kzg4844.Proof)(&proof)); err != nil {
	// 		return err
	// 	}
	// }

	// 3. check txlist comes from either calldata or blob, but not both exist
	calldataNotEmpty := len(g.Taiko.TxDataFromCalldata) != 0
	blobNotEmpty := len(g.Taiko.TxDataFromBlob) != 0
	if calldataNotEmpty && blobNotEmpty {
		return errors.New("txlist comes from either calldata or blob, but not both")
	}

	// 4. verify inputs length
	if len(g.Inputs) == 0 {
		return errors.New("no inputs")
	}
	if len(g.Inputs) > maxBlocksPerBatch {
		return fmt.Errorf(
			"too many inputs, expected at most %d, got %d",
			maxBlocksPerBatch,
			len(g.Inputs),
		)
	}

	// 5. verify the continuity of the blocks
	cur := g.Inputs[0].ParentHeader
	for input := range slices.Values(g.Inputs) {
		// check hash
		if cur.Hash() != input.Block.ParentHash() {
			return fmt.Errorf(
				"hash mismatch: expected %#x, got %#x",
				cur.Hash(),
				input.Block.ParentHash(),
			)
		}
		// check number
		if cur.Number.Uint64()+1 != input.Block.NumberU64() {
			return fmt.Errorf(
				"number mismatch: expected %d, got %d",
				cur.Number.Uint64()+1,
				input.Block.NumberU64(),
			)
		}
		// check state root
		if cur.Root != input.ParentHeader.Root {
			return fmt.Errorf(
				"state root mismatch: expected %#x, got %#x",
				cur.Root,
				input.ParentHeader.Root,
			)
		}
		cur = input.Block.Header()
	}
	return nil
}

func (g *BatchGuestInput) BlockMetadataFork() (BlockMetadataFork, error) {
	txListHash := keccak.Keccak(g.Taiko.TxDataFromCalldata)
	txsHash, err := g.calculatePacayaTxsHash(txListHash, g.Taiko.BatchProposed.BlobHashes())
	if err != nil {
		return nil, err
	}

	blocks := make([]pacaya.ITaikoInboxBlockParams, 0, len(g.Inputs))
	parentTs := g.Inputs[0].Block.Time()
	for input := range slices.Values(g.Inputs) {
		signalSlots, err := decodeAnchorV3Args_signalSlots(input.Taiko.AnchorTx.Data()[4:])
		if err != nil {
			return nil, err
		}
		if input.Block.Time() < parentTs || (input.Block.Time()-parentTs) > math.MaxUint8 {
			return nil, fmt.Errorf(
				"invalid delta block time, parent: %d, current: %d",
				parentTs,
				input.Block.Time(),
			)
		}
		blockParams := pacaya.ITaikoInboxBlockParams{
			NumTransactions: uint16(input.Block.Transactions().Len()) - 1,
			TimeShift:       uint8(input.Block.Time() - parentTs),
			SignalSlots:     signalSlots,
		}
		parentTs = input.Block.Time()
		blocks = append(blocks, blockParams)
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
	firstBlock := g.Inputs[0].Block
	lastBlock := g.Inputs[len(g.Inputs)-1].Block
	return &pacaya.ITaikoInboxTransition{
		ParentHash: firstBlock.ParentHash(),
		BlockHash:  lastBlock.Hash(),
		StateRoot:  lastBlock.Root(),
	}
}

func (g *BatchGuestInput) ForkVerifierAddress(proofType ProofType) common.Address {
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
