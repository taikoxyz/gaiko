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
		require.Equal(t, params.STDNetworkID, chainConfig.ChainID)
		require.Equal(t, core.STDOntakeBlock, chainConfig.OntakeBlock)
		require.Equal(t, core.STDPacayaBlock, chainConfig.PacayaBlock)
		require.NotNil(t, chainConfig.ShastaTime)
		require.Equal(t, uint64(1770987600), *chainConfig.ShastaTime)
	})

	t.Run("active_shasta_falls_back_to_std_default", func(t *testing.T) {
		chainSpec := &ChainSpec{
			Name: TaikoTransitionNetwork,
		}

		chainConfig, err := chainSpec.chainConfig(true)
		require.NoError(t, err)
		require.NotNil(t, chainConfig.ShastaTime)
		require.Equal(t, core.STDShastaTime, *chainConfig.ShastaTime)
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
