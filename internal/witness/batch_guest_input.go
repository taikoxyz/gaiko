package witness

import (
	"encoding/binary"
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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/manifest"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/pkg/utils"
)

var (
	_ GuestInput       = (*BatchGuestInput)(nil)
	_ json.Unmarshaler = (*BatchGuestInput)(nil)
)

type BatchGuestInput struct {
	Inputs []*SingleGuestInput
	Taiko  *TaikoGuestBatchInput
}

type TaikoGuestBatchInput struct {
	BatchID       uint64
	L1Header      *types.Header
	BatchProposed BlockProposed
	ChainSpec     *ChainSpec
	ProverData    *TaikoProverData
	DataSources   []*TaikoGuestDataSource
}

type TaikoGuestDataSource struct {
	TxDataFromCalldata []byte
	TxDataFromBlob     [][eth.BlobSize]byte
	BlobCommitments    *[][commitmentSize]byte
	BlobProofs         *[][proofSize]byte
	BlobProofType      BlobProofType
}

func (t *TaikoGuestBatchInput) primaryDataSource() *TaikoGuestDataSource {
	if len(t.DataSources) == 0 {
		return nil
	}
	return t.DataSources[0]
}

func (g *BatchGuestInput) GuestInputs() iter.Seq[*Pair] {
	return func(yield func(*Pair) bool) {
		switch g.Taiko.BatchProposed.HardFork() {
		case ShastaHardFork:
			g.yieldShastaGuestInputs(yield)
		default:
			g.yieldLegacyGuestInputs(yield)
		}
	}
}

func (g *BatchGuestInput) yieldLegacyGuestInputs(yield func(*Pair) bool) {
	batchProposed := g.Taiko.BatchProposed
	dataSource := g.Taiko.primaryDataSource()
	if dataSource == nil {
		log.Warn("missing taiko guest data source")
		return
	}

	var txs types.Transactions
	if batchProposed.BlobUsed() {
		var compressedTxListBuf []byte
		for _, blobDataBuf := range dataSource.TxDataFromBlob {
			blob := eth.Blob(blobDataBuf)
			if data, err := blob.ToData(); err != nil {
				log.Error("Parse blob data failed", "err", err)
			} else {
				compressedTxListBuf = append(compressedTxListBuf, data...)
			}
		}
		if len(compressedTxListBuf) == 0 {
			log.Warn("empty compressed tx list from blob data")
		}
		batchID := new(big.Int).SetUint64(g.Taiko.BatchID)
		var err error
		compressedTxListBuf, err = sliceTxList(
			batchID,
			compressedTxListBuf,
			batchProposed.BlobTxSliceParam(),
		)
		if err != nil {
			log.Warn(
				"Invalid txlist offset and size in metadata",
				"batchId", batchID,
				"err", err,
			)
		}
		txs = decompressTxList(
			compressedTxListBuf,
			blobMaxTxListBytes,
			batchProposed.BlobUsed(),
		)
	} else {
		txs = decompressTxList(
			dataSource.TxDataFromCalldata,
			calldataMaxTxListBytes,
			batchProposed.BlobUsed(),
		)
	}

	blockParams := batchProposed.BlockParams()
	if len(blockParams) == 0 {
		log.Warn("no block params available for batch guest input", "hardFork", batchProposed.HardFork())
		return
	}

	start := 0
	for i, blockParam := range blockParams {
		numTxs := int(blockParam.NumTransactions)
		end := min(start+numTxs, len(txs))
		if i >= len(g.Inputs) {
			log.Warn("insufficient inputs for block params", "index", i, "inputs", len(g.Inputs))
			break
		}
		extra := end - start
		if extra < 0 {
			extra = 0
		}
		blockTxs := make(types.Transactions, 0, 1+extra)
		blockTxs = append(blockTxs, g.Inputs[i].Taiko.AnchorTx)
		if start < end && end <= len(txs) {
			blockTxs = append(blockTxs, txs[start:end]...)
		}
		if !yield(&Pair{g.Inputs[i], blockTxs}) {
			return
		}
		start = end
	}
}

func (g *BatchGuestInput) yieldShastaGuestInputs(yield func(*Pair) bool) {
	shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
	if !ok {
		log.Warn("unexpected block proposed type for shasta batch input", "type", fmt.Sprintf("%T", g.Taiko.BatchProposed))
		return
	}
	eventData := shastaBlock.EventData()
	if eventData == nil {
		log.Warn("missing shasta event data")
		return
	}
	if len(g.Taiko.DataSources) == 0 {
		log.Warn("missing shasta data sources")
		return
	}
	if len(eventData.Derivation.Sources) != len(g.Taiko.DataSources) {
		log.Warn(
			"shasta derivation sources and data sources mismatch",
			"derivationSources", len(eventData.Derivation.Sources),
			"dataSources", len(g.Taiko.DataSources),
		)
	}

	var allBlockTxs []types.Transactions
	for idx, dataSource := range g.Taiko.DataSources {
		if idx >= len(eventData.Derivation.Sources) {
			log.Warn("extra data source without derivation metadata", "index", idx)
			break
		}
		combined, err := combineBlobData(dataSource.TxDataFromBlob)
		if err != nil {
			log.Error("failed to combine shasta blob data", "index", idx, "err", err)
			return
		}

		if len(combined) == 0 {
			log.Warn("shasta data source empty blob payload", "index", idx)
			continue
		}

		offset := int(eventData.Derivation.Sources[idx].BlobSlice.Offset)
		if offset+64 > len(combined) {
			log.Warn("shasta data source offset out of range", "index", idx, "offset", offset, "length", len(combined))
			return
		}

		version := combined[offset : offset+32]
		if version[31] != 1 {
			log.Warn("unexpected shasta manifest version", "index", idx, "version", version[31])
		}

		sizeBytes := combined[offset+32 : offset+64]
		size := binary.BigEndian.Uint64(sizeBytes[24:])
		end := offset + 64 + int(size)
		if end > len(combined) {
			log.Warn("shasta data source size out of range", "index", idx, "size", size, "length", len(combined))
			return
		}
		payload := combined[offset+64 : end]
		decoded, err := utils.Decompress(payload)
		if err != nil {
			log.Error("failed to decompress shasta manifest payload", "index", idx, "err", err)
			return
		}

		var blockTxs []types.Transactions
		var source manifest.DerivationSourceManifest
		if err := rlp.DecodeBytes(decoded, &source); err != nil {
			log.Error("failed to decode shasta derivation source manifest", "index", idx, "err", err)
			return
		}
		for _, block := range source.Blocks {
			blockTxs = append(blockTxs, block.Transactions)
		}
		allBlockTxs = append(allBlockTxs, blockTxs...)
	}

	if len(allBlockTxs) != len(g.Inputs) {
		log.Warn(
			"shasta manifest block count mismatch",
			"manifestBlocks", len(allBlockTxs),
			"inputs", len(g.Inputs),
		)
	}

	for i, input := range g.Inputs {
		var manifestTxs types.Transactions
		if i < len(allBlockTxs) {
			manifestTxs = allBlockTxs[i]
		}
		blockTxs := make(types.Transactions, 0, 1+len(manifestTxs))
		blockTxs = append(blockTxs, input.Taiko.AnchorTx)
		blockTxs = append(blockTxs, manifestTxs...)
		if !yield(&Pair{input, blockTxs}) {
			return
		}
	}
}

func combineBlobData(blobs [][eth.BlobSize]byte) ([]byte, error) {
	if len(blobs) == 0 {
		return nil, nil
	}
	var combined []byte
	for _, blobData := range blobs {
		blob := eth.Blob(blobData)
		data, err := blob.ToData()
		if err != nil {
			return nil, err
		}
		combined = append(combined, data...)
	}
	return combined, nil
}

func (g *BatchGuestInput) BlockProposed() BlockProposed {
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

	if g.Taiko == nil {
		return errors.New("missing taiko batch input")
	}
	if len(g.Taiko.DataSources) == 0 {
		return errors.New("missing taiko batch data sources")
	}

	// 2. validate blob commitments/proofs per data source
	for idx, dataSource := range g.Taiko.DataSources {
		blobProofType := getBlobProofType(proofType, dataSource.BlobProofType)
		switch blobProofType {
		case KzgVersionedHash:
			if len(dataSource.TxDataFromBlob) != 0 &&
				(dataSource.BlobCommitments == nil || len(dataSource.TxDataFromBlob) != len(*dataSource.BlobCommitments)) {
				return fmt.Errorf(
					"invalid blob commitments length in data source %d, expected: %d, got: %d",
					idx,
					len(dataSource.TxDataFromBlob),
					func() int {
						if dataSource.BlobCommitments == nil {
							return 0
						}
						return len(*dataSource.BlobCommitments)
					}(),
				)
			}
		case ProofOfEquivalence:
			if len(dataSource.TxDataFromBlob) != 0 &&
				(dataSource.BlobProofs == nil || len(dataSource.TxDataFromBlob) != len(*dataSource.BlobProofs)) {
				return fmt.Errorf(
					"invalid blob proofs length in data source %d, expected: %d, got: %d",
					idx,
					len(dataSource.TxDataFromBlob),
					func() int {
						if dataSource.BlobProofs == nil {
							return 0
						}
						return len(*dataSource.BlobProofs)
					}(),
				)
			}
		}

		// 3. check txlist comes from either calldata or blob, but not both exist
		calldataNotEmpty := len(dataSource.TxDataFromCalldata) != 0
		blobNotEmpty := len(dataSource.TxDataFromBlob) != 0
		if calldataNotEmpty && blobNotEmpty {
			return fmt.Errorf("data source %d txlist comes from either calldata or blob, but not both", idx)
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

func (g *BatchGuestInput) BlockMetadata() (BlockMetadata, error) {
	// Shasta uses a different metadata structure
	if g.Taiko.BatchProposed.IsShasta() {
		shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
		if !ok {
			return nil, fmt.Errorf("expected ShastaBlockProposed, got %T", g.Taiko.BatchProposed)
		}
		eventData := shastaBlock.EventData()
		if eventData == nil {
			return nil, errors.New("missing shasta event data")
		}
		return NewShastaBlockMetadata(eventData.Proposal.DerivationHash), nil
	}

	dataSource := g.Taiko.primaryDataSource()
	if dataSource == nil {
		return nil, errors.New("missing taiko data source")
	}
	txListHash := keccak.Keccak(dataSource.TxDataFromCalldata)
	txsHash, err := g.calculatePacayaTxsHash(txListHash, g.Taiko.BatchProposed.BlobHashes())
	if err != nil {
		return nil, err
	}

	blocks := make([]pacaya.ITaikoInboxBlockParams, 0, len(g.Inputs))
	parentTS := g.Inputs[0].Block.Time()

	if len(g.Inputs) != len(g.Taiko.BatchProposed.BlockParams()) {
		return nil, fmt.Errorf(
			"mismatched inputs: %d and block parameters: %d length",
			len(g.Inputs),
			len(g.Taiko.BatchProposed.BlockParams()),
		)
	}
	var signalSlots [][32]byte
	for idx, input := range g.Inputs {
		signalSlots, err = decodeAnchorV3ArgsSignalSlots(input.Taiko.AnchorTx.Data()[4:])
		if err != nil {
			return nil, err
		}
		if input.Block.Time() < parentTS || (input.Block.Time()-parentTS) > math.MaxUint8 {
			return nil, fmt.Errorf(
				"invalid delta block time, parent: %d, current: %d",
				parentTS,
				input.Block.Time(),
			)
		}
		blockParams := pacaya.ITaikoInboxBlockParams{
			NumTransactions: g.Taiko.BatchProposed.BlockParams()[idx].NumTransactions,
			TimeShift:       uint8(input.Block.Time() - parentTS),
			SignalSlots:     signalSlots,
		}
		parentTS = input.Block.Time()
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
	// Shasta uses a different transition structure
	if g.Taiko.BatchProposed.IsShasta() {
		return g.buildShastaTransitions()
	}

	firstBlock := g.Inputs[0].Block
	lastBlock := g.Inputs[len(g.Inputs)-1].Block
	return &pacaya.ITaikoInboxTransition{
		ParentHash: firstBlock.ParentHash(),
		BlockHash:  lastBlock.Hash(),
		StateRoot:  lastBlock.Root(),
	}
}

func (g *BatchGuestInput) buildShastaTransitions() []common.Hash {
	shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
	if !ok {
		return nil
	}
	eventData := shastaBlock.EventData()
	if eventData == nil {
		return nil
	}

	// Compute proposal hash once
	proposalHash := hashProposal(&eventData.Proposal)

	// Build transitions for each block
	transitionHashes := make([]common.Hash, 0, len(g.Inputs))

	// Get parent transition hash and checkpoint from prover_data if available
	var (
		parentTransitionHash common.Hash
		checkpoint           *ShastaProposalCheckpoint
		designatedProver     common.Address
		actualProver         common.Address
		designatedProvided   bool
	)

	if g.Taiko.ProverData != nil {
		if g.Taiko.ProverData.ParentTransitionHash != nil {
			parentTransitionHash = *g.Taiko.ProverData.ParentTransitionHash
		}
		checkpoint = g.Taiko.ProverData.Checkpoint
		designatedProver = g.Taiko.ProverData.DesignatedProver
		designatedProvided = g.Taiko.ProverData.designatedProverSet
		actualProver = g.Taiko.ProverData.ActualProver
	}

	// If no prover data, fall back to core state and proposer
	if parentTransitionHash == (common.Hash{}) {
		parentTransitionHash = eventData.CoreState.LastFinalizedTransitionHash
	}
	if !designatedProvided {
		designatedProver = eventData.Proposal.Proposer
	}
	if actualProver == (common.Address{}) {
		actualProver = designatedProver
	}

	for i, input := range g.Inputs {
		block := input.Block

		// Use checkpoint from prover_data if available, otherwise construct from block
		var shastaCheckpoint ShastaCheckpoint
		if checkpoint != nil && i == len(g.Inputs)-1 {
			// Use prover_data checkpoint for the last block
			shastaCheckpoint = ShastaCheckpoint{
				BlockNumber: checkpoint.BlockNumber,
				BlockHash:   checkpoint.BlockHash,
				StateRoot:   checkpoint.StateRoot,
			}
		} else {
			// Construct from block data
			shastaCheckpoint = ShastaCheckpoint{
				BlockNumber: block.NumberU64(),
				BlockHash:   block.Hash(),
				StateRoot:   block.Root(),
			}
		}

		// Create transition
		transition := &ShastaTransition{
			ProposalHash:         proposalHash,
			ParentTransitionHash: parentTransitionHash,
			Checkpoint:           shastaCheckpoint,
		}

		// Create metadata
		metadata := &ShastaTransitionMetadata{
			DesignatedProver: designatedProver,
			ActualProver:     actualProver,
		}

		// Compute transition hash
		transitionHash := hashTransitionWithMetadata(transition, metadata)
		transitionHashes = append(transitionHashes, transitionHash)

		// Update parent for next iteration
		parentTransitionHash = transitionHash
	}

	return transitionHashes
}

func (g *BatchGuestInput) ForkVerifierAddress(proofType ProofType) common.Address {
	// Use the last block's timestamp for fork activation check
	var timestamp uint64
	if len(g.Inputs) > 0 {
		timestamp = g.Inputs[len(g.Inputs)-1].Block.Time()
	}
	return g.Taiko.ChainSpec.getForkVerifierAddress(
		g.Taiko.BatchProposed.BlockNumber(),
		timestamp,
		proofType,
	)
}

func (g *BatchGuestInput) Prover() common.Address {
	return g.Taiko.ProverData.ActualProver
}

func (g *BatchGuestInput) ChainID() uint64 {
	return g.Taiko.ChainSpec.ChainID
}

func (g *BatchGuestInput) ID() ID {
	return ID{
		BatchID: g.Taiko.BatchID,
	}
}

func (g *BatchGuestInput) IsTaiko() bool {
	return g.Taiko.ChainSpec.IsTaiko
}

func (g *BatchGuestInput) ChainConfig() (*params.ChainConfig, error) {
	// Dynamically set ShastaTime based on the input data
	// Only activate Shasta fork if the block proposed data is actually using Shasta
	activeShasta := g.Taiko.BatchProposed.IsShasta()
	return g.Taiko.ChainSpec.chainConfig(activeShasta)
}
