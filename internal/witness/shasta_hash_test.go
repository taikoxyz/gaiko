package witness

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestHashProposal(t *testing.T) {
	proposal := &ShastaProposal{
		ID:                             12345,
		Timestamp:                      193_828_690,
		EndOfSubmissionWindowTimestamp: 193_829_690,
		Proposer:                       common.HexToAddress("0x1234567890AbcdEF1234567890aBcdef12345678"),
		ParentProposalHash:             common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		OriginBlockNumber:              73_826,
		OriginBlockHash:                common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		BasefeeSharingPctg:             42,
		Sources: []ShastaDerivationSource{
			{
				IsForcedInclusion: true,
				BlobSlice: ShastaBlobSlice{
					BlobHashes: []common.Hash{
						common.HexToHash("0x67890abcdef1234567890abcdef123451234567890abcdef1234567890abcdef"),
					},
					Offset:    0,
					Timestamp: 100,
				},
			},
			{
				IsForcedInclusion: false,
				BlobSlice: ShastaBlobSlice{
					BlobHashes: []common.Hash{
						common.HexToHash("0x567890abcdef123451234567890abcdef123456767890abcdef1234890abcdef"),
					},
					Offset:    100,
					Timestamp: 200,
				},
			},
		},
	}

	proposalHash := hashProposal(proposal)
	expected := common.HexToHash("0x13af2d05799894db3462512e3ecf5ae8877b80b1e2db3963654ac70f6dd49f88")
	assert.Equal(t, expected, proposalHash, "proposal hash mismatch")
}

func TestHashCommitment(t *testing.T) {
	commitment := &ShastaCommitment{
		FirstProposalID:              42,
		FirstProposalParentBlockHash: common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000999"),
		LastProposalHash:             common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000123456"),
		ActualProver:                 common.HexToAddress("0x0000000000000000000000000000000000012345"),
		EndBlockNumber:               1000,
		EndStateRoot:                 common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000abcdef"),
		Transitions: []ShastaTransition{
			{
				Proposer:  common.HexToAddress("0x0000000000000000000000000000000000001111"),
				Timestamp: 123_456_789,
				BlockHash: common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000003333"),
			},
		},
	}

	commitmentHash := hashCommitment(commitment)
	expected := common.HexToHash("0x079961e990a2be01ebe286ee2fdd382fde2349730971fe32a821da9dec67559e")
	assert.Equal(t, expected, commitmentHash, "commitment hash mismatch")
}
