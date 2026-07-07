package witness

import (
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func TestTransitionChainConfig(t *testing.T) {
	t.Run("active_shasta_uses_chain_spec_timestamp", func(t *testing.T) {
		chainSpec := &ChainSpec{
			Name: TaikoTransitionNetwork,
			HardForks: HardForks{
				{
					SpecID:    SpecID(PacayaHardFork),
					Condition: BlockNumber(0),
				},
				{
					SpecID:    SpecID(ShastaHardFork),
					Condition: BlockTimestamp(1770987600),
				},
			},
		}

		chainConfig, err := chainSpec.chainConfig(true)
		require.NoError(t, err)
		require.Equal(t, params.TaikoInternalNetworkID, chainConfig.ChainID)
		require.Equal(t, core.InternalDevnetOntakeBlock, chainConfig.OntakeBlock)
		require.Equal(t, core.InternalDevnetPacayaBlock, chainConfig.PacayaBlock)
		require.NotNil(t, chainConfig.ShastaTime)
		require.Equal(t, uint64(1770987600), *chainConfig.ShastaTime)
	})

	t.Run("active_shasta_falls_back_to_internal_default", func(t *testing.T) {
		chainSpec := &ChainSpec{
			Name: TaikoTransitionNetwork,
		}

		chainConfig, err := chainSpec.chainConfig(true)
		require.NoError(t, err)
		require.NotNil(t, chainConfig.ShastaTime)
		require.Equal(t, core.InternalShastaTime, *chainConfig.ShastaTime)
	})

	t.Run("inactive_shasta_clears_shasta_time", func(t *testing.T) {
		chainSpec := &ChainSpec{
			Name: TaikoTransitionNetwork,
		}

		chainConfig, err := chainSpec.chainConfig(false)
		require.NoError(t, err)
		require.Nil(t, chainConfig.ShastaTime)
	})
}

func TestVerifyChainSpecRejectsNameMismatch(t *testing.T) {
	var mainnetSpec *ChainSpec
	for _, chainSpec := range defaultSupportedChainSpecs {
		if chainSpec.Name == TaikoMainnetNetwork {
			mainnetSpec = chainSpec
			break
		}
	}
	require.NotNil(t, mainnetSpec)

	mismatchedSpec := *mainnetSpec
	mismatchedSpec.Name = TaikoDevNetwork

	err := defaultSupportedChainSpecs.verifyChainSpec(&mismatchedSpec)
	require.EqualError(t, err, "unexpected name")
}
