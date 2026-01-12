package witness

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/tests/fixtures"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/pkg/utils"
)

func TestShastaManifestMatchesInputBlockParams(t *testing.T) {
	payload, err := fixtures.ReadShastaFixture("input-52.json")
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

	offset := int(eventData.Proposal.Sources[0].BlobSlice.Offset)
	require.GreaterOrEqual(t, len(combined), offset+64)

	sizeBytes := combined[offset+32 : offset+64]
	size := binary.BigEndian.Uint64(sizeBytes[24:])
	end := offset + 64 + int(size)
	require.LessOrEqual(t, end, len(combined))

	decoded, err := utils.Decompress(combined[offset+64 : end])
	require.NoError(t, err)

	m, err := decodeShastaDerivationSourceManifest(decoded)
	require.NoError(t, err)
	require.Len(t, m.Blocks, 1)

	require.True(t, validateInputBlockParam(m.Blocks[0], input.Inputs[0].Block))
}

func TestShastaGuestInputsDoesNotFallbackToDefaultManifest(t *testing.T) {
	payload, err := fixtures.ReadShastaFixture("input-52.json")
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

func TestShastaAnchorLinkageDecodesCheckpoint(t *testing.T) {
	payload, err := fixtures.ReadShastaFixture("input-52.json")
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

	require.True(t, validateShastaBlockBaseFee([]*SingleGuestInput{input}, false, grandparent))
}

func TestCalcNextShastaBaseFee_RaikoVector(t *testing.T) {
	result := calcNextShastaBaseFee(
		16_000_000,
		15_956_512,
		5_000_000,
		240,
		shastaDefaultElasticityMultiplier,
		shastaDefaultBaseFeeDenominator,
	)
	require.Equal(t, uint64(5_059_102), result)
}
