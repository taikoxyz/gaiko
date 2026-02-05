package witness

import (
	"bytes"
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
	"github.com/ethereum/go-ethereum/consensus/taiko"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
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

var shastaDefaultManifestObserver func(anchorBlockNumber uint64, isForceInclusion bool)

type BatchGuestInput struct {
	Inputs []*SingleGuestInput
	Taiko  *TaikoGuestBatchInput
}

type TaikoGuestBatchInput struct {
	BatchID             uint64
	L1Header            *types.Header
	L1AncestorHeaders   []*types.Header
	L2GrandparentHeader *types.Header
	BatchProposed       BlockProposed
	ChainSpec           *ChainSpec
	ProverData          *TaikoProverData
	DataSources         []*TaikoGuestDataSource
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
	if g.Taiko.ProverData == nil {
		log.Warn("missing shasta prover data")
		return
	}
	if len(eventData.Proposal.Sources) != len(g.Taiko.DataSources) {
		log.Warn(
			"shasta derivation sources and data sources mismatch",
			"derivationSources", len(eventData.Proposal.Sources),
			"dataSources", len(g.Taiko.DataSources),
		)
	}

	if err := g.validateShastaBlockTimestamp(); err != nil {
		log.Warn("shasta block timestamp validation failed", "err", err)
		return
	}

	if len(g.Inputs) == 0 {
		log.Warn("missing shasta inputs")
		return
	}

	lastParentBlockTimestamp := g.Inputs[0].ParentHeader.Time
	lastParentBlockGasLimit := g.Inputs[0].ParentHeader.GasLimit
	proposalTimestamp := eventData.Proposal.Timestamp
	forkTimestamp := shastaForkTimestamp(g.Taiko.ChainSpec)
	isGenesisParent := g.Inputs[0].ParentHeader.Number.Uint64() == 0
	useInitBaseFee := isGenesisParent

	var allBlockTxs []types.Transactions
	for idx, dataSource := range g.Taiko.DataSources {
		if idx >= len(eventData.Proposal.Sources) {
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
			start, size, ok := shastaBlobTxSliceParamForSource(eventData.Proposal.Sources[idx], combined)
			if !ok {
				return nil
			}
			payload := combined[start : start+size]
			d, err := utils.Decompress(payload)
			if err != nil {
				return nil
			}
			return d
		}()

		source, decodeErr := decodeShastaDerivationSourceManifest(decoded)

		var validManifest *manifest.DerivationSourceManifest

		if idx == len(g.Taiko.DataSources)-1 {
			// Normal source
			if decodeErr == nil && validateNormalProposalManifest(g, source, g.Taiko.ProverData.LastAnchorBlockNumber) {
				if !validateShastaBlockBaseFee(g.Inputs, useInitBaseFee, g.Taiko.L2GrandparentHeader) {
					log.Warn("shasta block base fee is invalid, use default manifest")
					timestamp := clampTimestampLowerBound(lastParentBlockTimestamp, proposalTimestamp, forkTimestamp)
					coinbase := g.Taiko.BatchProposed.Proposer()
					anchorBlockNumber := g.Taiko.ProverData.LastAnchorBlockNumber
					validManifest = g.createDefaultManifest(timestamp, coinbase, anchorBlockNumber, lastParentBlockGasLimit, isGenesisParent)
				} else {
					validManifest = source
				}
			} else {
				// Fallback
				timestamp := clampTimestampLowerBound(lastParentBlockTimestamp, proposalTimestamp, forkTimestamp)
				coinbase := g.Taiko.BatchProposed.Proposer()
				anchorBlockNumber := g.Taiko.ProverData.LastAnchorBlockNumber
				validManifest = g.createDefaultManifest(timestamp, coinbase, anchorBlockNumber, lastParentBlockGasLimit, isGenesisParent)
			}
		} else {
			// Force inclusion source
			timestamp := clampTimestampLowerBound(lastParentBlockTimestamp, proposalTimestamp, forkTimestamp)
			coinbase := g.Taiko.BatchProposed.Proposer()
			anchorBlockNumber := g.Taiko.ProverData.LastAnchorBlockNumber
			forceManifest := g.createDefaultManifest(timestamp, coinbase, anchorBlockNumber, lastParentBlockGasLimit, isGenesisParent)
			if decodeErr == nil && validateForceIncProposalManifest(source) && len(source.Blocks) > 0 {
				forceManifest.Blocks[0].Transactions = source.Blocks[0].Transactions
			} else {
				if shastaDefaultManifestObserver != nil {
					shastaDefaultManifestObserver(anchorBlockNumber, true)
				}
			}
			validManifest = forceManifest
			if len(validManifest.Blocks) > 0 {
				lastParentBlockTimestamp = validManifest.Blocks[0].Timestamp
				lastParentBlockGasLimit = validManifest.Blocks[0].GasLimit
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

// shastaProposalMaxBlocks is the maximum number of blocks allowed in a Shasta proposal.
// Keep this aligned with Raiko's `PROPOSAL_MAX_BLOCKS`.
const shastaProposalMaxBlocks = 192

const shastaBlobDataPrefixSize = 64

func shastaBlobTxSliceParamForSource(
	source ShastaDerivationSource,
	combined []byte,
) (int, int, bool) {
	if len(source.BlobSlice.BlobHashes) == 0 {
		return 0, 0, false
	}

	offset := int(source.BlobSlice.Offset)
	if offset > eth.BlobSize-shastaBlobDataPrefixSize {
		return 0, 0, false
	}
	if offset+shastaBlobDataPrefixSize > len(combined) {
		return 0, 0, false
	}

	version := combined[offset : offset+32]
	if !isShastaManifestVersion(version) {
		return 0, 0, false
	}

	sizeBytes := combined[offset+32 : offset+64]
	size := binary.BigEndian.Uint64(sizeBytes[24:])
	start := offset + shastaBlobDataPrefixSize
	remaining := len(combined) - start
	if size > uint64(remaining) {
		return 0, 0, false
	}
	end := start + int(size)
	if end > len(combined) {
		return 0, 0, false
	}
	return start, int(size), true
}

func isShastaManifestVersion(version []byte) bool {
	if len(version) != 32 {
		return false
	}
	if version[31] != 1 {
		return false
	}
	for i := 0; i < 31; i++ {
		if version[i] != 0 {
			return false
		}
	}
	return true
}

type legacyDerivationSourceManifest struct {
	Blocks []*manifest.BlockManifest
}

func decodeShastaDerivationSourceManifest(data []byte) (*manifest.DerivationSourceManifest, error) {
	if len(data) == 0 {
		return nil, errors.New("empty shasta derivation source manifest")
	}

	stream := rlp.NewStream(bytes.NewReader(data), 0)
	if _, err := stream.List(); err != nil {
		return nil, err
	}

	// New format: `DerivationSourceManifest` is encoded as `[proverAuthBytes, blocks]`.
	// Legacy format (used by current fixtures and raiko): `DerivationSourceManifest` is encoded as
	// `[blocks]`, without `proverAuthBytes`.
	kind, _, err := stream.Kind()
	if err != nil {
		return nil, err
	}

	if kind == rlp.List {
		blocks, err := decodeShastaBlockManifestsWithLimit(stream)
		if err != nil {
			return nil, err
		}
		if err := stream.ListEnd(); err != nil {
			return nil, err
		}
		return &manifest.DerivationSourceManifest{
			ProverAuthBytes: nil,
			Blocks:          blocks,
		}, nil
	}

	var proverAuthBytes []byte
	if err := stream.Decode(&proverAuthBytes); err != nil {
		return nil, err
	}
	blocks, err := decodeShastaBlockManifestsWithLimit(stream)
	if err != nil {
		return nil, err
	}
	if err := stream.ListEnd(); err != nil {
		return nil, err
	}
	return &manifest.DerivationSourceManifest{
		ProverAuthBytes: proverAuthBytes,
		Blocks:          blocks,
	}, nil
}

func decodeShastaBlockManifestsWithLimit(stream *rlp.Stream) ([]*manifest.BlockManifest, error) {
	if _, err := stream.List(); err != nil {
		return nil, err
	}

	blocks := make([]*manifest.BlockManifest, 0, min(shastaProposalMaxBlocks, 16))
	for stream.MoreDataInList() {
		if len(blocks) >= shastaProposalMaxBlocks {
			return nil, fmt.Errorf("shasta proposal block number exceeds limit %d", shastaProposalMaxBlocks)
		}
		var block manifest.BlockManifest
		if err := stream.Decode(&block); err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}
	if err := stream.ListEnd(); err != nil {
		return nil, err
	}
	return blocks, nil
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

	// 2. resolve expected blob hashes per data source
	expectedBlobHashesPerSource := make([][]common.Hash, 0, len(g.Taiko.DataSources))
	if g.Taiko.BatchProposed.IsShasta() {
		shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
		if !ok {
			return fmt.Errorf("expected ShastaBlockProposed, got %T", g.Taiko.BatchProposed)
		}
		eventData := shastaBlock.EventData()
		if eventData == nil {
			return errors.New("missing shasta event data")
		}
		if len(eventData.Proposal.Sources) != len(g.Taiko.DataSources) {
			return fmt.Errorf(
				"shasta derivation sources and data sources mismatch: expected %d, got %d",
				len(eventData.Proposal.Sources),
				len(g.Taiko.DataSources),
			)
		}
		for _, source := range eventData.Proposal.Sources {
			expectedBlobHashesPerSource = append(expectedBlobHashesPerSource, source.BlobSlice.BlobHashes)
		}
	} else {
		blobHashes := g.Taiko.BatchProposed.BlobHashes()
		hashes := make([]common.Hash, len(blobHashes))
		for i, hash := range blobHashes {
			hashes[i] = common.Hash(hash)
		}
		for range g.Taiko.DataSources {
			expectedBlobHashesPerSource = append(expectedBlobHashesPerSource, hashes)
		}
	}
	if len(expectedBlobHashesPerSource) != len(g.Taiko.DataSources) {
		return fmt.Errorf(
			"data sources length mismatch: expected %d, got %d",
			len(expectedBlobHashesPerSource),
			len(g.Taiko.DataSources),
		)
	}

	// 3. validate blob commitments/proofs per data source
	for idx, dataSource := range g.Taiko.DataSources {
		// check txlist comes from either calldata or blob, but not both exist
		calldataNotEmpty := len(dataSource.TxDataFromCalldata) != 0
		blobNotEmpty := len(dataSource.TxDataFromBlob) != 0
		if calldataNotEmpty && blobNotEmpty {
			return fmt.Errorf("data source %d txlist comes from either calldata or blob, but not both", idx)
		}

		expectedBlobHashes := expectedBlobHashesPerSource[idx]
		if len(expectedBlobHashes) != len(dataSource.TxDataFromBlob) {
			return fmt.Errorf(
				"source %d blob hashes length mismatch, expected: %d, got: %d",
				idx,
				len(expectedBlobHashes),
				len(dataSource.TxDataFromBlob),
			)
		}

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
			if len(dataSource.TxDataFromBlob) == 0 {
				break
			}
			commitments := dataSource.BlobCommitments
			if commitments == nil {
				return fmt.Errorf("missing blob commitments in data source %d", idx)
			}
			for i, blobData := range dataSource.TxDataFromBlob {
				commitment := kzg4844.Commitment((*commitments)[i])
				expectedHash := expectedBlobHashes[i]
				if eth.KZGToVersionedHash(commitment) != expectedHash {
					return fmt.Errorf("versioned hash mismatch in data source %d, index %d", idx, i)
				}
				blob := eth.Blob(blobData)
				if err := verifyBlob(blobProofType, &blob, commitment, nil); err != nil {
					return err
				}
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
			if len(dataSource.TxDataFromBlob) == 0 {
				break
			}
			commitments := dataSource.BlobCommitments
			proofs := dataSource.BlobProofs
			if commitments == nil {
				return fmt.Errorf("missing blob commitments in data source %d", idx)
			}
			if proofs == nil {
				return fmt.Errorf("missing blob proofs in data source %d", idx)
			}
			for i, blobData := range dataSource.TxDataFromBlob {
				commitment := kzg4844.Commitment((*commitments)[i])
				proof := kzg4844.Proof((*proofs)[i])
				expectedHash := expectedBlobHashes[i]
				if eth.KZGToVersionedHash(commitment) != expectedHash {
					return fmt.Errorf("versioned hash mismatch in data source %d, index %d", idx, i)
				}
				blob := eth.Blob(blobData)
				if err := verifyBlob(blobProofType, &blob, commitment, &proof); err != nil {
					return err
				}
			}
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

	// 5. Shasta-specific anchor linkage and origin checks
	if g.Taiko.BatchProposed.IsShasta() {
		shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
		if !ok {
			return fmt.Errorf("expected ShastaBlockProposed, got %T", g.Taiko.BatchProposed)
		}
		eventData := shastaBlock.EventData()
		if eventData == nil {
			return errors.New("missing shasta event data")
		}
		if g.Taiko.L1Header == nil {
			return errors.New("missing l1 header for shasta batch input")
		}
		if eventData.Proposal.OriginBlockNumber != g.Taiko.L1Header.Number.Uint64() {
			return fmt.Errorf(
				"l1 origin block number mismatch, expected %d, got %d",
				eventData.Proposal.OriginBlockNumber,
				g.Taiko.L1Header.Number.Uint64(),
			)
		}
		if eventData.Proposal.OriginBlockHash != g.Taiko.L1Header.Hash() {
			return fmt.Errorf(
				"l1 origin block hash mismatch, expected %#x, got %#x",
				eventData.Proposal.OriginBlockHash,
				g.Taiko.L1Header.Hash(),
			)
		}
		if _, err := encodeShastaProposal(&eventData.Proposal); err != nil {
			return err
		}
		if err := verifyShastaAnchorLinkage(g.Inputs, g.Taiko.L1AncestorHeaders, eventData.Proposal.OriginBlockHash); err != nil {
			return err
		}
		if g.Taiko.ProverData != nil && g.Taiko.ProverData.Checkpoint != nil && len(g.Inputs) > 0 {
			lastBlock := g.Inputs[len(g.Inputs)-1].Block
			expectedCheckpoint := ShastaCheckpoint{
				BlockNumber: lastBlock.NumberU64(),
				BlockHash:   lastBlock.Hash(),
				StateRoot:   lastBlock.Root(),
			}
			if expectedCheckpoint.BlockNumber != g.Taiko.ProverData.Checkpoint.BlockNumber ||
				expectedCheckpoint.BlockHash != g.Taiko.ProverData.Checkpoint.BlockHash ||
				expectedCheckpoint.StateRoot != g.Taiko.ProverData.Checkpoint.StateRoot {
				return fmt.Errorf("shasta checkpoint mismatch, expected %+v, got %+v", expectedCheckpoint, g.Taiko.ProverData.Checkpoint)
			}
		}
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
		return NewShastaBlockMetadata(&eventData.Proposal), nil
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

func (g *BatchGuestInput) buildShastaTransition() TransitionInputData {
	shastaBlock, ok := g.Taiko.BatchProposed.(*ShastaBlockProposed)
	if !ok || shastaBlock.EventData() == nil || len(g.Inputs) == 0 {
		return TransitionInputData{}
	}
	eventData := shastaBlock.EventData()

	lastBlock := g.Inputs[len(g.Inputs)-1].Block
	checkpoint := ShastaCheckpoint{
		BlockNumber: lastBlock.NumberU64(),
		BlockHash:   lastBlock.Hash(),
		StateRoot:   lastBlock.Root(),
	}

	proposalHash := hashProposal(&eventData.Proposal)
	parentBlockHash := g.Inputs[0].ParentHeader.Hash()
	actualProver := common.Address{}
	if g.Taiko.ProverData != nil {
		actualProver = g.Taiko.ProverData.ActualProver
	}

	return TransitionInputData{
		ProposalID:         eventData.Proposal.ID,
		ProposalHash:       proposalHash,
		ParentProposalHash: eventData.Proposal.ParentProposalHash,
		ParentBlockHash:    parentBlockHash,
		ActualProver:       actualProver,
		Transition: ShastaTransitionInput{
			Proposer:  eventData.Proposal.Proposer,
			Timestamp: eventData.Proposal.Timestamp,
		},
		Checkpoint: checkpoint,
	}
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
	// Shasta rule constants synced with raiko.
	shastaBlockGasLimitMaxChange uint64 = 200
	shastaGasLimitDenominator    uint64 = 1_000_000
	shastaMaxBlockGasLimitBase   uint64 = 45_000_000
	shastaMinBlockGasLimitBase   uint64 = 10_000_000

	shastaTimestampMaxOffset          uint64 = 12 * 128
	shastaBlockTimeTarget             uint64 = 2
	shastaMaxGasTargetTargetPercent   uint64 = 95
	shastaMinBaseFee                  uint64 = 5_000_000
	shastaMaxBaseFee                  uint64 = 1_000_000_000
	shastaDefaultBaseFeeDenominator   uint64 = 8
	shastaDefaultElasticityMultiplier uint64 = 2
)

func (g *BatchGuestInput) validateShastaBlockTimestamp() error {
	proposalTimestamp := g.Taiko.BatchProposed.ProposedAt()
	forkTimestamp := shastaForkTimestamp(g.Taiko.ChainSpec)
	for _, input := range g.Inputs {
		blockTimestamp := input.Block.Time()

		if blockTimestamp > proposalTimestamp {
			return fmt.Errorf("block timestamp %d exceeds proposal timestamp %d", blockTimestamp, proposalTimestamp)
		}

		parentTimestamp := input.ParentHeader.Time
		lowerBound := clampTimestampLowerBound(parentTimestamp, proposalTimestamp, forkTimestamp)

		if blockTimestamp < lowerBound {
			return fmt.Errorf("block timestamp %d is less than calculated lower bound %d", blockTimestamp, lowerBound)
		}
	}
	return nil
}

func validAnchorInNormalProposal(
	blocks []*manifest.BlockManifest,
	lastAnchorBlockNumber uint64,
	l1OriginBlockNumber uint64,
) bool {
	minAnchor := saturatingSub(l1OriginBlockNumber, manifest.AnchorMaxOffset)
	maxAnchor := l1OriginBlockNumber

	hasAnchorGrow := false
	var prevAnchor uint64
	prevAnchorSet := false
	for _, block := range blocks {
		anchor := block.AnchorBlockNumber
		if anchor < lastAnchorBlockNumber {
			log.Error(
				"anchor below last anchor block number",
				"anchor", anchor,
				"lastAnchorBlockNumber", lastAnchorBlockNumber,
			)
			return false
		}
		if anchor > lastAnchorBlockNumber {
			hasAnchorGrow = true
		}
		if prevAnchorSet && anchor < prevAnchor {
			log.Error("anchor is not in order", "blocks", blocks)
			return false
		}
		prevAnchor = anchor
		prevAnchorSet = true
		if anchor < minAnchor || anchor > maxAnchor {
			log.Error(
				"anchor out of range",
				"anchor", anchor,
				"min", minAnchor,
				"max", maxAnchor,
			)
			return false
		}
	}

	if !hasAnchorGrow {
		log.Error("anchor is not growing", "lastAnchorBlockNumber", lastAnchorBlockNumber)
	}
	return hasAnchorGrow
}

func validateNormalProposalManifest(
	input *BatchGuestInput,
	m *manifest.DerivationSourceManifest,
	lastAnchorBlockNumber uint64,
) bool {
	if len(m.Blocks) > shastaProposalMaxBlocks {
		log.Error(
			"manifest block number exceeds limit",
			"count", len(m.Blocks),
			"limit", shastaProposalMaxBlocks,
		)
		return false
	}
	proposalBlockNumber := input.Taiko.BatchProposed.BlockNumber()
	l1OriginBlockNumber := saturatingSub(proposalBlockNumber, 1)
	if !validAnchorInNormalProposal(m.Blocks, lastAnchorBlockNumber, l1OriginBlockNumber) {
		log.Error("valid_anchor_in_proposal failed", "lastAnchorBlockNumber", lastAnchorBlockNumber)
		return false
	}
	if !validateShastaBlockGasLimit(m.Blocks, input.Inputs) {
		log.Error("validate_shasta_block_gas_limit failed")
		return false
	}
	if !validateShastaManifestBlockTimestamp(m.Blocks, input) {
		log.Error("validate_shasta_block_timesatmp failed")
		return false
	}
	return true
}

func validateForceIncProposalManifest(m *manifest.DerivationSourceManifest) bool {
	if len(m.Blocks) != 1 {
		log.Error("force inclusion manifest must have exactly 1 block", "count", len(m.Blocks))
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
	if manifestBlock.GasLimit+taiko.AnchorV3V4GasLimit != inputBlock.GasLimit() {
		log.Error("gas limit mismatch", "manifest", manifestBlock.GasLimit, "input", inputBlock.GasLimit())
		return false
	}
	return true
}

func validateShastaBlockGasLimit(
	manifestBlocks []*manifest.BlockManifest,
	blockGuestInputs []*SingleGuestInput,
) bool {
	if len(blockGuestInputs) == 0 {
		return false
	}
	parentGasLimit := blockGuestInputs[0].ParentHeader.GasLimit
	if blockGuestInputs[0].ParentHeader.Number.Uint64() != 0 && parentGasLimit >= taiko.AnchorV3V4GasLimit {
		parentGasLimit -= taiko.AnchorV3V4GasLimit
	}
	maxBlockGasLimit := shastaMaxBlockGasLimitBase
	minBlockGasLimit := shastaMinBlockGasLimitBase
	for _, manifestBlock := range manifestBlocks {
		blockGasLimit := manifestBlock.GasLimit
		upperLimit := min(
			maxBlockGasLimit,
			parentGasLimit*(shastaGasLimitDenominator+shastaBlockGasLimitMaxChange)/shastaGasLimitDenominator,
		)
		lowerLimit := min(
			max(
				minBlockGasLimit,
				parentGasLimit*(shastaGasLimitDenominator-shastaBlockGasLimitMaxChange)/shastaGasLimitDenominator,
			),
			upperLimit,
		)
		if blockGasLimit < lowerLimit || blockGasLimit > upperLimit {
			log.Error(
				"block gas limit out of bounds",
				"blockGasLimit", blockGasLimit,
				"lowerLimit", lowerLimit,
				"upperLimit", upperLimit,
			)
			return false
		}
		parentGasLimit = blockGasLimit
	}
	return true
}

func validateShastaManifestBlockTimestamp(
	blocks []*manifest.BlockManifest,
	batchInput *BatchGuestInput,
) bool {
	if len(batchInput.Inputs) == 0 {
		return false
	}
	proposalTimestamp := batchInput.Taiko.BatchProposed.ProposedAt()
	forkTimestamp := shastaForkTimestamp(batchInput.Taiko.ChainSpec)
	parentTimestamp := batchInput.Inputs[0].ParentHeader.Time
	for _, block := range blocks {
		blockTimestamp := block.Timestamp
		if blockTimestamp > proposalTimestamp {
			log.Error(
				"block timestamp exceeds proposal timestamp",
				"blockTimestamp", blockTimestamp,
				"proposalTimestamp", proposalTimestamp,
			)
			return false
		}
		lowerBound := clampTimestampLowerBound(parentTimestamp, proposalTimestamp, forkTimestamp)
		if blockTimestamp < lowerBound {
			log.Error(
				"block timestamp below lower bound",
				"blockTimestamp", blockTimestamp,
				"lowerBound", lowerBound,
			)
			return false
		}
		parentTimestamp = blockTimestamp
	}
	return true
}

func shastaForkTimestamp(chainSpec *ChainSpec) uint64 {
	if chainSpec == nil {
		return 0
	}
	for _, fork := range chainSpec.HardForks {
		if fork.SpecID != SpecID(ShastaHardFork) {
			continue
		}
		if timestamp, ok := fork.Condition.(BlockTimestamp); ok {
			return uint64(timestamp)
		}
		return 0
	}
	return 0
}

func clampTimestampLowerBound(parentTimestamp uint64, proposalTimestamp uint64, shastaForkTimestamp uint64) uint64 {
	lowerBound := saturatingAdd(parentTimestamp, 1)
	if proposalTimestamp > shastaTimestampMaxOffset {
		altLowerBound := proposalTimestamp - shastaTimestampMaxOffset
		if altLowerBound > lowerBound {
			lowerBound = altLowerBound
		}
	}
	if shastaForkTimestamp > lowerBound {
		lowerBound = shastaForkTimestamp
	}
	return lowerBound
}

func saturatingSub(a uint64, b uint64) uint64 {
	if a <= b {
		return 0
	}
	return a - b
}

func saturatingMul(a uint64, b uint64) uint64 {
	if a == 0 || b == 0 {
		return 0
	}
	if a > math.MaxUint64/b {
		return math.MaxUint64
	}
	return a * b
}

func saturatingAdd(a uint64, b uint64) uint64 {
	if a > math.MaxUint64-b {
		return math.MaxUint64
	}
	return a + b
}

func clampShastaBaseFee(baseFee uint64) uint64 {
	if baseFee < shastaMinBaseFee {
		return shastaMinBaseFee
	}
	if baseFee > shastaMaxBaseFee {
		return shastaMaxBaseFee
	}
	return baseFee
}

func calcNextShastaBaseFee(
	parentGasLimit uint64,
	parentGasUsed uint64,
	parentBaseFee uint64,
	parentBlockTime uint64,
	elasticityMultiplier uint64,
	baseFeeChangeDenominator uint64,
) uint64 {
	if elasticityMultiplier == 0 {
		return clampShastaBaseFee(parentBaseFee)
	}
	parentGasTarget := parentGasLimit / elasticityMultiplier
	if parentGasTarget == 0 {
		return clampShastaBaseFee(parentBaseFee)
	}

	adjustedTarget1 := saturatingMul(parentGasTarget, parentBlockTime) / shastaBlockTimeTarget
	adjustedTarget2 := saturatingMul(parentGasLimit, shastaMaxGasTargetTargetPercent) / 100
	parentAdjustedGasTarget := min(adjustedTarget1, adjustedTarget2)

	if parentGasUsed == parentAdjustedGasTarget {
		return clampShastaBaseFee(parentBaseFee)
	}

	if parentGasUsed > parentAdjustedGasTarget {
		gasUsedDelta := parentGasUsed - parentAdjustedGasTarget
		adjustment := saturatingMul(parentBaseFee, gasUsedDelta) / parentGasTarget / baseFeeChangeDenominator
		if adjustment < 1 {
			return clampShastaBaseFee(saturatingAdd(parentBaseFee, 1))
		}
		return clampShastaBaseFee(saturatingAdd(parentBaseFee, adjustment))
	}

	gasUsedDelta := parentAdjustedGasTarget - parentGasUsed
	adjustment := saturatingMul(parentBaseFee, gasUsedDelta) / parentGasTarget / baseFeeChangeDenominator
	if adjustment > parentBaseFee {
		return clampShastaBaseFee(0)
	}
	return clampShastaBaseFee(parentBaseFee - adjustment)
}

func validateShastaBlockBaseFee(
	blockGuestInputs []*SingleGuestInput,
	useInitBaseFee bool,
	l2GrandparentHeader *types.Header,
) bool {
	if len(blockGuestInputs) == 0 {
		return false
	}
	firstBaseFee := blockGuestInputs[0].Block.BaseFee()
	if firstBaseFee == nil {
		return false
	}
	if useInitBaseFee {
		if firstBaseFee.Uint64() != params.ShastaInitialBaseFee {
			log.Warn("shasta block base fee is invalid",
				"expected", params.ShastaInitialBaseFee,
				"found", firstBaseFee.Uint64(),
			)
			return false
		}
	} else {
		parentBaseFee := blockGuestInputs[0].ParentHeader.BaseFee
		if parentBaseFee == nil {
			return false
		}
		parentBlockTime := uint64(shastaBlockTimeTarget)
		if l2GrandparentHeader != nil {
			parentBlockTime = saturatingSub(blockGuestInputs[0].ParentHeader.Time, l2GrandparentHeader.Time)
		}
		expectedBaseFee := calcNextShastaBaseFee(
			blockGuestInputs[0].ParentHeader.GasLimit,
			blockGuestInputs[0].ParentHeader.GasUsed,
			parentBaseFee.Uint64(),
			parentBlockTime,
			shastaDefaultElasticityMultiplier,
			shastaDefaultBaseFeeDenominator,
		)
		if expectedBaseFee != firstBaseFee.Uint64() {
			return false
		}
	}

	for i := 1; i < len(blockGuestInputs); i++ {
		block := blockGuestInputs[i].Block
		actualBaseFee := block.BaseFee()
		if actualBaseFee == nil {
			return false
		}
		prevBlock := blockGuestInputs[i-1].Block
		prevBaseFee := prevBlock.BaseFee()
		if prevBaseFee == nil {
			return false
		}
		parentBlockTime := saturatingSub(prevBlock.Time(), blockGuestInputs[i-1].ParentHeader.Time)
		expectedBaseFee := calcNextShastaBaseFee(
			prevBlock.GasLimit(),
			prevBlock.GasUsed(),
			prevBaseFee.Uint64(),
			parentBlockTime,
			shastaDefaultElasticityMultiplier,
			shastaDefaultBaseFeeDenominator,
		)
		if expectedBaseFee != actualBaseFee.Uint64() {
			return false
		}
	}
	return true
}

type shastaAnchorParam struct {
	blockNumber uint64
	blockHash   common.Hash
	stateRoot   common.Hash
}

func verifyShastaAnchorLinkage(
	inputs []*SingleGuestInput,
	l1AncestorHeaders []*types.Header,
	expectedParentHash common.Hash,
) error {
	if len(inputs) == 0 {
		return errors.New("missing shasta inputs")
	}
	anchorParamSet := make(map[shastaAnchorParam]struct{})
	for _, input := range inputs {
		if input.Taiko.AnchorTx == nil {
			return errors.New("missing shasta anchor tx")
		}
		data := input.Taiko.AnchorTx.Data()
		if len(data) < 4 {
			return errors.New("invalid shasta anchor tx data")
		}
		checkpoint, err := decodeShastaAnchorCheckpoint(data[4:])
		if err != nil {
			return err
		}
		anchorParamSet[shastaAnchorParam{
			blockNumber: checkpoint.BlockNumber,
			blockHash:   checkpoint.BlockHash,
			stateRoot:   checkpoint.StateRoot,
		}] = struct{}{}
	}

	if len(l1AncestorHeaders) == 0 {
		return errors.New("l1 ancestor headers is empty")
	}

	l1AncestorSet := make(map[shastaAnchorParam]struct{})
	lastParentHash := l1AncestorHeaders[0].Hash()
	l1AncestorSet[shastaAnchorParam{
		blockNumber: l1AncestorHeaders[0].Number.Uint64(),
		blockHash:   lastParentHash,
		stateRoot:   l1AncestorHeaders[0].Root,
	}] = struct{}{}
	for _, header := range l1AncestorHeaders[1:] {
		if header.ParentHash != lastParentHash {
			return fmt.Errorf(
				"l1 ancestor header parent hash mismatch, expected: %#x, got: %#x",
				lastParentHash,
				header.ParentHash,
			)
		}
		currHash := header.Hash()
		l1AncestorSet[shastaAnchorParam{
			blockNumber: header.Number.Uint64(),
			blockHash:   currHash,
			stateRoot:   header.Root,
		}] = struct{}{}
		lastParentHash = currHash
	}

	for anchorParam := range anchorParamSet {
		if _, ok := l1AncestorSet[anchorParam]; !ok {
			return fmt.Errorf("anchor param not found in l1 ancestor hash set: %+v", anchorParam)
		}
	}

	if lastParentHash != expectedParentHash {
		return fmt.Errorf(
			"l1 ancestor hash mismatch, expected: %#x, got: %#x",
			expectedParentHash,
			lastParentHash,
		)
	}
	return nil
}

func (g *BatchGuestInput) createDefaultManifest(
	timestamp uint64,
	coinbase common.Address,
	anchorBlockNumber uint64,
	gasLimit uint64,
	isGenesisParent bool,
) *manifest.DerivationSourceManifest {
	if !isGenesisParent && gasLimit >= taiko.AnchorV3V4GasLimit {
		gasLimit -= taiko.AnchorV3V4GasLimit
	}
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
