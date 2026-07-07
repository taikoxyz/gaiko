package witness

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/tests/fixtures"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/pkg/utils"
)

const shastaFixtureName = "input-3.json"

func TestShastaManifestMatchesInputBlockParams(t *testing.T) {
	payload, err := fixtures.ReadShastaFixture(shastaFixtureName)
	require.NoError(t, err)

	var input BatchGuestInput
	require.NoError(t, json.Unmarshal(payload, &input))
	require.Len(t, input.Inputs, 1)

	shastaBlock, ok := input.Taiko.BatchProposed.(*ShastaBlockProposed)
	require.True(t, ok)
	eventData := shastaBlock.EventData()
	require.NotNil(t, eventData)
	require.Len(t, eventData.Proposal.Sources, 1)
	require.Len(t, input.Taiko.DataSources, 1)

	combined, err := combineBlobData(input.Taiko.DataSources[0].TxDataFromBlob)
	require.NoError(t, err)

	start, size, ok := shastaBlobTxSliceParamForSource(eventData.Proposal.Sources[0], combined)
	require.True(t, ok)

	decoded, err := utils.Decompress(combined[start : start+size])
	require.NoError(t, err)

	m, err := decodeShastaDerivationSourceManifest(decoded)
	require.NoError(t, err)
	require.Len(t, m.Blocks, 1)

	require.True(t, validateInputBlockParam(m.Blocks[0], input.Inputs[0].Block))
}

func TestShastaGuestInputsDoesNotFallbackToDefaultManifest(t *testing.T) {
	payload, err := fixtures.ReadShastaFixture(shastaFixtureName)
	require.NoError(t, err)

	var input BatchGuestInput
	require.NoError(t, json.Unmarshal(payload, &input))
	require.Len(t, input.Inputs, 1)

	count := 0
	for range input.GuestInputs() {
		count++
	}
	require.Equal(t, len(input.Inputs), count)
}

func TestShastaDefaultManifest_ForceInclusionUsesLastAnchor(t *testing.T) {
	var observedAnchor uint64
	shastaDefaultManifestObserver = func(anchorBlockNumber uint64, isForceInclusion bool) {
		if isForceInclusion {
			observedAnchor = anchorBlockNumber
		}
	}
	defer func() {
		shastaDefaultManifestObserver = nil
	}()

	input := BatchGuestInput{
		Inputs: []*SingleGuestInput{makeShastaGuestInputWithAnchorTx(20, 21, 200, 30_000_000, false)},
		Taiko: &TaikoGuestBatchInput{
			BatchID: 1,
			BatchProposed: NewShastaBlockProposed(&ShastaEventData{
				Proposal: ShastaProposal{
					Timestamp:          200,
					Proposer:           common.Address{0x11},
					OriginBlockNumber:  99,
					OriginBlockHash:    common.Hash{0x22},
					ParentProposalHash: common.Hash{0x33},
					Sources: []ShastaDerivationSource{
						{
							IsForcedInclusion: true,
							BlobSlice: ShastaBlobSlice{
								Offset: 0,
							},
						},
						{
							IsForcedInclusion: false,
							BlobSlice: ShastaBlobSlice{
								Offset: 0,
							},
						},
					},
				},
			}),
			ChainSpec: &ChainSpec{
				Name:      TaikoDevNetwork,
				ChainID:   1,
				MaxSpecID: SpecID(ShastaHardFork),
				HardForks: HardForks{{SpecID: SpecID(ShastaHardFork), Condition: BlockNumber(0)}},
				Eip1559Constants: &Eip1559Constants{
					BaseFeeChangeDenominator:      big.NewInt(8),
					BaseFeeMaxIncreaseDenominator: big.NewInt(8),
					BaseFeeMaxDecreaseDenominator: big.NewInt(8),
					ElasticityMultiplier:          big.NewInt(2),
				},
				L1Contract:           map[SpecID]*common.Address{},
				VerifierAddressForks: map[SpecID]VerifierAddressFork{},
				GenesisTime:          0,
				SecondsPerSlot:       12,
				IsTaiko:              true,
			},
			ProverData: &TaikoProverData{
				LastAnchorBlockNumber: 42,
			},
			DataSources: []*TaikoGuestDataSource{
				{IsForcedInclusion: true, TxDataFromBlob: emptyManifestBlob()},
				{IsForcedInclusion: false, TxDataFromBlob: emptyManifestBlob()},
			},
		},
	}

	collectGuestInputs(&input)
	require.Equal(t, uint64(42), observedAnchor)
}

func TestShastaAnchorLinkageDecodesCheckpoint(t *testing.T) {
	payload, err := fixtures.ReadShastaFixture(shastaFixtureName)
	require.NoError(t, err)

	var input BatchGuestInput
	require.NoError(t, json.Unmarshal(payload, &input))
	require.Len(t, input.Inputs, 1)

	shastaBlock, ok := input.Taiko.BatchProposed.(*ShastaBlockProposed)
	require.True(t, ok)
	eventData := shastaBlock.EventData()
	require.NotNil(t, eventData)

	require.NoError(t, verifyShastaAnchorLinkage(
		input.Inputs,
		input.Taiko.L1AncestorHeaders,
		eventData.Proposal.OriginBlockHash,
	))
}

func TestShastaAnchorLinkageAllowsEmptyAncestors(t *testing.T) {
	input := makeShastaGuestInputWithAnchorTx(1, 2, 200, 30_000_000, false)
	input.Taiko.AnchorTx = types.NewTx(&types.LegacyTx{Data: make([]byte, 4+96)})

	require.NoError(t, verifyShastaAnchorLinkage(
		[]*SingleGuestInput{input},
		nil,
		common.Hash{},
	))
}

func makeShastaGuestInput(parentNumber uint64, blockNumber uint64, timestamp uint64, gasLimit uint64) *SingleGuestInput {
	return makeShastaGuestInputWithAnchorTx(parentNumber, blockNumber, timestamp, gasLimit, true)
}

func makeShastaGuestInputWithAnchorTx(
	parentNumber uint64,
	blockNumber uint64,
	timestamp uint64,
	gasLimit uint64,
	withAnchorTx bool,
) *SingleGuestInput {
	parentHeader := &types.Header{
		Number:   new(big.Int).SetUint64(parentNumber),
		Time:     timestamp - 1,
		GasLimit: gasLimit,
		GasUsed:  gasLimit - 1,
		BaseFee:  big.NewInt(5_000_000),
	}
	blockHeader := &types.Header{
		Number:     new(big.Int).SetUint64(blockNumber),
		ParentHash: parentHeader.Hash(),
		Time:       timestamp,
		GasLimit:   gasLimit,
		GasUsed:    gasLimit - 2,
		BaseFee:    big.NewInt(5_000_000),
	}
	block := types.NewBlockWithHeader(blockHeader)

	var anchorTx *types.Transaction
	if withAnchorTx {
		anchorTx = types.NewTx(&types.LegacyTx{Data: make([]byte, 4)})
	}

	return &SingleGuestInput{
		Block:        block,
		ParentHeader: parentHeader,
		ChainSpec: &ChainSpec{
			Name:      TaikoDevNetwork,
			ChainID:   1,
			MaxSpecID: SpecID(ShastaHardFork),
			HardForks: HardForks{{SpecID: SpecID(ShastaHardFork), Condition: BlockNumber(0)}},
			Eip1559Constants: &Eip1559Constants{
				BaseFeeChangeDenominator:      big.NewInt(8),
				BaseFeeMaxIncreaseDenominator: big.NewInt(8),
				BaseFeeMaxDecreaseDenominator: big.NewInt(8),
				ElasticityMultiplier:          big.NewInt(2),
			},
			L1Contract:           map[SpecID]*common.Address{},
			VerifierAddressForks: map[SpecID]VerifierAddressFork{},
			GenesisTime:          0,
			SecondsPerSlot:       12,
			IsTaiko:              true,
		},
		Taiko: &TaikoGuestInput{
			AnchorTx: anchorTx,
		},
	}
}

func emptyManifestBlob() [][eth.BlobSize]byte {
	var blob [eth.BlobSize]byte
	return [][eth.BlobSize]byte{blob}
}

func collectGuestInputs(input *BatchGuestInput) []struct {
	input *SingleGuestInput
	txs   types.Transactions
} {
	pairs := make([]struct {
		input *SingleGuestInput
		txs   types.Transactions
	}, 0)
	for pair := range input.GuestInputs() {
		pairs = append(pairs, struct {
			input *SingleGuestInput
			txs   types.Transactions
		}{
			input: pair.Input,
			txs:   pair.Txs,
		})
	}
	return pairs
}

func TestValidateShastaBlockBaseFee_UsesGrandparent(t *testing.T) {
	parentHeader := &types.Header{
		GasLimit: 16_000_000,
		GasUsed:  15_956_512,
		BaseFee:  big.NewInt(5_000_000),
		Time:     240,
	}
	blockHeader := &types.Header{
		GasLimit: 16_000_000,
		GasUsed:  15_956_512,
		BaseFee:  big.NewInt(5_059_102),
		Time:     241,
	}
	block := types.NewBlockWithHeader(blockHeader)
	input := &SingleGuestInput{Block: block, ParentHeader: parentHeader}
	grandparent := &types.Header{Time: 0}

	require.True(t, validateShastaBlockBaseFee([]*SingleGuestInput{input}, false, grandparent, shastaMinBaseFee))
}

func TestCalcNextShastaBaseFee_RaikoVector(t *testing.T) {
	result := calcNextShastaBaseFee(
		16_000_000,
		15_956_512,
		5_000_000,
		240,
		shastaDefaultElasticityMultiplier,
		shastaDefaultBaseFeeDenominator,
		shastaMinBaseFee,
	)
	require.Equal(t, uint64(5_059_102), result)
}

func TestCalcNextShastaBaseFee_SaturatingMul(t *testing.T) {
	result := calcNextShastaBaseFee(
		12_431_977_784_026_273_569,
		183_807_743_286_270_785,
		15_468_160,
		12_597_304_404_566_764_638,
		shastaDefaultElasticityMultiplier,
		shastaDefaultBaseFeeDenominator,
		shastaMinBaseFee,
	)
	require.Equal(t, uint64(15_468_160), result)
}

func TestCalcNextShastaBaseFee_SaturatingAdd(t *testing.T) {
	result := calcNextShastaBaseFee(
		math.MaxUint64,
		math.MaxUint64/2+1,
		math.MaxUint64,
		2,
		shastaDefaultElasticityMultiplier,
		shastaDefaultBaseFeeDenominator,
		shastaMinBaseFee,
	)
	require.Equal(t, uint64(1_000_000_000), result)
}

func TestCalcNextShastaBaseFee_MainnetMinBaseFee(t *testing.T) {
	result := calcNextShastaBaseFee(
		16_000_000,
		8_000_000,
		6_000_000,
		shastaBlockTimeTarget,
		shastaDefaultElasticityMultiplier,
		shastaDefaultBaseFeeDenominator,
		shastaMainnetMinBaseFee,
	)
	require.Equal(t, shastaMainnetMinBaseFee, result)
}

func TestShastaChainSpecificOffsets(t *testing.T) {
	mainnetID := params.TaikoMainnetNetworkID.Uint64()
	transitionID := uint64(167014)
	otherID := uint64(167001)

	require.Equal(t, shastaMainnetAnchorMaxOffset, shastaAnchorMaxOffsetForChain(mainnetID))
	require.Equal(t, shastaMainnetAnchorMaxOffset, shastaAnchorMaxOffsetForChain(transitionID))
	require.Equal(t, shastaAnchorMaxOffset, shastaAnchorMaxOffsetForChain(otherID))

	require.Equal(t, shastaMainnetTimestampMaxOffset, shastaTimestampMaxOffsetForChain(mainnetID))
	require.Equal(t, shastaMainnetTimestampMaxOffset, shastaTimestampMaxOffsetForChain(transitionID))
	require.Equal(t, shastaHoodiTimestampMaxOffset, shastaTimestampMaxOffsetForChain(otherID))

	require.Equal(t, shastaMainnetMinBaseFee, shastaMinBaseFeeForChain(mainnetID))
	require.Equal(t, shastaMainnetMinBaseFee, shastaMinBaseFeeForChain(transitionID))
	require.Equal(t, shastaMinBaseFee, shastaMinBaseFeeForChain(otherID))
}
