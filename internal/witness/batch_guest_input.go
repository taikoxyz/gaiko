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
	IsForcedInclusion  bool
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

	if err := g.validateShastaBlockTimestamp(); err != nil {
		log.Warn("shasta block timestamp validation failed", "err", err)
		return
	}

	var allBlockTxs []types.Transactions
	for idx, dataSource := range g.Taiko.DataSources {
		if idx >= len(eventData.Derivation.Sources) {
			log.Warn("extra data source without derivation metadata", "index", idx)
			break
		}

		if idx == len(g.Taiko.DataSources)-1 {
			if dataSource.IsForcedInclusion {
				log.Warn("last source should be normal source", "index", idx)
				return
			}
		} else {
			if !dataSource.IsForcedInclusion {
				log.Warn("begin sources should be force inclusion source", "index", idx)
				return
			}
		}

		combined, err := combineBlobData(dataSource.TxDataFromBlob)
		if err != nil {
			log.Error("failed to combine shasta blob data", "index", idx, "err", err)
			return
		}

		decoded := func() []byte {
			if len(combined) == 0 {
				return nil
			}
			offset := int(eventData.Derivation.Sources[idx].BlobSlice.Offset)
			if offset+64 > len(combined) {
				return nil
			}
			version := combined[offset : offset+32]
			if version[31] != 1 {
				log.Warn("unexpected shasta manifest version", "index", idx, "version", version[31])
			}
			sizeBytes := combined[offset+32 : offset+64]
			size := binary.BigEndian.Uint64(sizeBytes[24:])
			end := offset + 64 + int(size)
			if end > len(combined) {
				return nil
			}
			payload := combined[offset+64 : end]
			d, err := utils.Decompress(payload)
			if err != nil {
				return nil
			}
			return d
		}()

		var source manifest.DerivationSourceManifest
		decodeErr := rlp.DecodeBytes(decoded, &source)

		var validManifest *manifest.DerivationSourceManifest

		if idx == len(g.Taiko.DataSources)-1 {
			// Normal source
			if decodeErr == nil && validateNormalProposalManifest(&source, g.Taiko.ProverData.LastAnchorBlockNumber) {
				validManifest = &source
			} else {
				// Fallback
				timestamp := g.Taiko.L1Header.Time
				coinbase := g.Taiko.BatchProposed.Proposer()
				anchorBlockNumber := g.Taiko.ProverData.LastAnchorBlockNumber
				lastInput := g.Inputs[len(g.Inputs)-1]
				gasLimit := lastInput.ParentHeader.GasLimit

				validManifest = g.createDefaultManifest(timestamp, coinbase, anchorBlockNumber, gasLimit)
			}
		} else {
			// Force inclusion source
			if decodeErr == nil && validateForceIncProposalManifest(&source) {
				validManifest = &source
			} else {
				// Fallback
				timestamp := g.Taiko.L1Header.Time
				coinbase := g.Taiko.BatchProposed.Proposer()
				anchorBlockNumber := uint64(0)
				gasLimit := uint64(0)

				validManifest = g.createDefaultManifest(timestamp, coinbase, anchorBlockNumber, gasLimit)
			}
		}

		var blockTxs []types.Transactions
		for i, block := range validManifest.Blocks {
			inputIdx := idx + i
			if inputIdx >= len(g.Inputs) {
				log.Warn("input index out of range", "index", inputIdx)
				break
			}
			if !validateInputBlockParam(block, g.Inputs[inputIdx].Block) {
				log.Warn("input block param validation failed", "index", inputIdx)
				return
			}
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
		return g.buildShastaTransition()
	}

	firstBlock := g.Inputs[0].Block
	lastBlock := g.Inputs[len(g.Inputs)-1].Block
	return &pacaya.ITaikoInboxTransition{
		ParentHash: firstBlock.ParentHash(),
		BlockHash:  lastBlock.Hash(),
		StateRoot:  lastBlock.Root(),
	}
}

func (g *BatchGuestInput) buildShastaTransition() common.Hash {
	shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
	if !ok {
		return common.Hash{}
	}
	eventData := shastaBlock.EventData()
	if eventData == nil {
		return common.Hash{}
	}

	if len(g.Inputs) == 0 {
		return common.Hash{}
	}

	var (
		proposalHash         = hashProposal(&eventData.Proposal)
		parentTransitionHash common.Hash
		checkpoint           *ShastaProposalCheckpoint
		designatedProver     common.Address
		designatedProverSet  bool
		actualProver         common.Address
	)

	if g.Taiko.ProverData != nil {
		if g.Taiko.ProverData.ParentTransitionHash != nil {
			parentTransitionHash = *g.Taiko.ProverData.ParentTransitionHash
		}
		checkpoint = g.Taiko.ProverData.Checkpoint
		designatedProver = g.Taiko.ProverData.DesignatedProver
		designatedProverSet = g.Taiko.ProverData.designatedProverSet
		actualProver = g.Taiko.ProverData.ActualProver
	}

	if checkpoint == nil {
		// If no checkpoint in prover_data, use the first block as checkpoint
		lastBlock := g.Inputs[len(g.Inputs)-1].Block
		checkpoint = &ShastaProposalCheckpoint{
			BlockNumber: lastBlock.NumberU64(),
			BlockHash:   lastBlock.Hash(),
			StateRoot:   lastBlock.Root(),
		}
	}

	// If no prover data, fall back to core state and proposer
	if parentTransitionHash == (common.Hash{}) {
		parentTransitionHash = eventData.CoreState.LastFinalizedTransitionHash
	}
	if !designatedProverSet {
		designatedProver = eventData.Proposal.Proposer
	}
	if actualProver == (common.Address{}) {
		actualProver = designatedProver
	}

	// Create transition
	transition := &ShastaTransition{
		ProposalHash:         proposalHash,
		ParentTransitionHash: parentTransitionHash,
		Checkpoint: ShastaCheckpoint{
			BlockNumber: checkpoint.BlockNumber,
			BlockHash:   checkpoint.BlockHash,
			StateRoot:   checkpoint.StateRoot,
		},
	}

	// Create metadata
	metadata := &ShastaTransitionMetadata{
		DesignatedProver: designatedProver,
		ActualProver:     actualProver,
	}

	// Compute transition hash
	transitionHash := hashTransitionWithMetadata(transition, metadata)
	return transitionHash
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

const (
	timestampMaxOffset = 12 * 32
	proposalMaxBlocks  = 384
)

func (g *BatchGuestInput) validateShastaBlockTimestamp() error {
	for _, input := range g.Inputs {
		blockTimestamp := input.Block.Time()
		proposalTimestamp := g.Taiko.BatchProposed.ProposedAt()

		if blockTimestamp > proposalTimestamp {
			return fmt.Errorf("block timestamp %d exceeds proposal timestamp %d", blockTimestamp, proposalTimestamp)
		}

		parentTimestamp := input.ParentHeader.Time
		lowerBound := parentTimestamp + 1
		if proposalTimestamp > timestampMaxOffset {
			altLowerBound := proposalTimestamp - timestampMaxOffset
			if altLowerBound > lowerBound {
				lowerBound = altLowerBound
			}
		}

		if blockTimestamp < lowerBound {
			return fmt.Errorf("block timestamp %d is less than calculated lower bound %d", blockTimestamp, lowerBound)
		}
	}
	return nil
}

func validAnchorInNormalProposal(blocks []*manifest.BlockManifest, lastAnchorBlockNumber uint64) bool {
	for _, block := range blocks {
		if block.AnchorBlockNumber > lastAnchorBlockNumber {
			return true
		}
	}
	return false
}

func validateNormalProposalManifest(m *manifest.DerivationSourceManifest, lastAnchorBlockNumber uint64) bool {
	if len(m.Blocks) > proposalMaxBlocks {
		log.Error("manifest block number exceeds limit", "count", len(m.Blocks), "limit", proposalMaxBlocks)
		return false
	}
	if !validAnchorInNormalProposal(m.Blocks, lastAnchorBlockNumber) {
		log.Error("valid_anchor_in_proposal failed", "lastAnchorBlockNumber", lastAnchorBlockNumber)
		return false
	}
	return true
}

func validateForceIncProposalManifest(m *manifest.DerivationSourceManifest) bool {
	if len(m.Blocks) != 1 {
		log.Error("force inclusion manifest must have exactly 1 block", "count", len(m.Blocks))
		return false
	}
	block := m.Blocks[0]
	if block.Timestamp != 0 || block.Coinbase != (common.Address{}) || block.AnchorBlockNumber != 0 || block.GasLimit != 0 {
		log.Error("invalid force inclusion block manifest", "block", block)
		return false
	}
	return true
}

func validateInputBlockParam(manifestBlock *manifest.BlockManifest, inputBlock *types.Block) bool {
	if manifestBlock.Timestamp != inputBlock.Time() {
		log.Error("timestamp mismatch", "manifest", manifestBlock.Timestamp, "input", inputBlock.Time())
		return false
	}
	if manifestBlock.Coinbase != inputBlock.Coinbase() {
		log.Error("coinbase mismatch", "manifest", manifestBlock.Coinbase, "input", inputBlock.Coinbase())
		return false
	}
	if manifestBlock.GasLimit != inputBlock.GasLimit() {
		log.Error("gas limit mismatch", "manifest", manifestBlock.GasLimit, "input", inputBlock.GasLimit())
		return false
	}
	return true
}

func (g *BatchGuestInput) createDefaultManifest(
	timestamp uint64,
	coinbase common.Address,
	anchorBlockNumber uint64,
	gasLimit uint64,
) *manifest.DerivationSourceManifest {
	return &manifest.DerivationSourceManifest{
		Blocks: []*manifest.BlockManifest{
			{
				Timestamp:         timestamp,
				Coinbase:          coinbase,
				AnchorBlockNumber: anchorBlockNumber,
				GasLimit:          gasLimit,
				Transactions:      types.Transactions{},
			},
		},
	}
}
