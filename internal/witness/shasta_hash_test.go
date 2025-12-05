package witness

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestHashProposal(t *testing.T) {
	// Test case: 3 values (packed, proposer, derivationHash) - no coreStateHash after PR #642
	proposal := &ShastaProposal{
		ID:                             3549,
		Timestamp:                      1761830468,
		EndOfSubmissionWindowTimestamp: 0,
		Proposer:                       common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc"),
		DerivationHash:                 common.HexToHash("0x85422bfec85e2cb6d5ca9f52858a74b680865c0134c0e29af710d8e01d58898a"),
	}

	proposalHash := hashProposal(proposal)
	expected := common.HexToHash("0x0fd2106121ee59690d5c49dcbd1603e9eedff34da6dd6afe5de01d30188d770d")

	assert.Equal(t, expected, proposalHash, "proposal hash mismatch")
}

func TestHashCheckpoint(t *testing.T) {
	checkpoint := &ShastaCheckpoint{
		BlockNumber: 1512,
		BlockHash:   common.HexToHash("0x83cf1bb221b330d372ce0fbca82cb060fa028d3f6bfd62a74197789e25ac2b5f"),
		StateRoot:   common.HexToHash("0x63651766d70b5aaf0320fc63421f4d1fdf6fe828514e21e05615e9c2f93c9c7d"),
	}

	checkpointHash := hashCheckpoint(checkpoint)
	t.Logf("Checkpoint hash: %s", checkpointHash.Hex())
	// We don't have expected value from Raiko for checkpoint alone, but log it for debugging
}

func TestHashTransitionWithMetadata(t *testing.T) {
	// Test case from Raiko's test_shasta_transition_hash
	transition := &ShastaTransition{
		ProposalHash:         common.HexToHash("0xd469fc0c500db1c87cd4fcf0650628cf4be84b03feb29dbca9ce1daee2750274"),
		ParentTransitionHash: common.HexToHash("0x66aa40046aa64a8e0a7ecdbbc70fb2c63ebdcb2351e7d0b626ed3cb4f55fb388"),
		Checkpoint: ShastaCheckpoint{
			BlockNumber: 1512,
			BlockHash:   common.HexToHash("0x83cf1bb221b330d372ce0fbca82cb060fa028d3f6bfd62a74197789e25ac2b5f"),
			StateRoot:   common.HexToHash("0x63651766d70b5aaf0320fc63421f4d1fdf6fe828514e21e05615e9c2f93c9c7d"),
		},
	}

	metadata := &ShastaTransitionMetadata{
		DesignatedProver: common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc"),
		ActualProver:     common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc"),
	}

	transitionHash := hashTransitionWithMetadata(transition, metadata)
	t.Logf("Transition hash: %s", transitionHash.Hex())
	// Expected value from Raiko test (continuation of test in Raiko source)
}
