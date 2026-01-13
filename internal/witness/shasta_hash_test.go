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

func TestHashPublicInput(t *testing.T) {
	proveInputHash := common.HexToHash("0xb836ee1f972e8bcd4766bede4a9fa5267d8b6ec7cd6088562aca0b07b15f57bc")
	chainID := uint64(167001)
	verifier := common.HexToAddress("0x00f9f60C79e38c08b785eE4F1a849900693C6630")
	got := hashPublicInput(proveInputHash, chainID, verifier, common.Address{})
	expected := common.HexToHash("0x6d0ea3eb338aa3e2d85b21394d3ea426574ab7764726376a5364dee132fcd3d7")
	assert.Equal(t, expected, got, "public input hash mismatch")
}

// TestShastaAggregationOutput tests the shasta aggregation output hash matching Rust test.
func TestShastaAggregationOutput(t *testing.T) {
	commitment := &ShastaCommitment{
		FirstProposalID:              12345,
		FirstProposalParentBlockHash: common.Hash{},
		LastProposalHash:             common.Hash{},
		ActualProver:                 common.HexToAddress("0x1111111111111111111111111111111111111111"),
		EndBlockNumber:               1,
		EndStateRoot:                 common.Hash{},
		Transitions:                  []ShastaTransition{},
	}
	chainID := uint64(167001)
	verifier := common.HexToAddress("0x00f9f60C79e38c08b785eE4F1a849900693C6630")
	sgxInstance := common.HexToAddress("0xdc95623058E847fA38e56a0Fa466Bf52C48eFA32")

	commitmentHash := hashCommitment(commitment)
	finalHash := hashPublicInput(commitmentHash, chainID, verifier, sgxInstance)

	expected := common.HexToHash("0x5ffd635c42c7e6f7a5aa6c83be7db37dd1c24f1b474606ef0901b9b32beffaae")
	assert.Equal(t, expected, finalHash, "shasta aggregation output hash mismatch")
}

func TestPCDOrderMatters(t *testing.T) {
	chainID := uint64(167001)
	verifier := common.HexToAddress("0x00f9f60C79e38c08b785eE4F1a849900693C6630")
	actualProver := common.HexToAddress("0x0000000000000000000000000000000000001111")
	proposer := common.HexToAddress("0x0000000000000000000000000000000000002222")

	pcd1 := ProofCarryData{
		ChainID:  chainID,
		Verifier: verifier,
		TransitionInput: TransitionInputData{
			ProposalID:         1,
			ProposalHash:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000101"),
			ParentProposalHash: common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
			ParentBlockHash:    common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000aa"),
			ActualProver:       actualProver,
			Transition: ShastaTransitionInput{
				Proposer:  proposer,
				Timestamp: 100,
			},
			Checkpoint: ShastaCheckpoint{
				BlockNumber: 10,
				BlockHash:   common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000bb"),
				StateRoot:   common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000cc"),
			},
		},
	}

	pcd2 := ProofCarryData{
		ChainID:  chainID,
		Verifier: verifier,
		TransitionInput: TransitionInputData{
			ProposalID:         2,
			ProposalHash:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000202"),
			ParentProposalHash: pcd1.TransitionInput.ProposalHash,
			ParentBlockHash:    pcd1.TransitionInput.Checkpoint.BlockHash,
			ActualProver:       actualProver,
			Transition: ShastaTransitionInput{
				Proposer:  proposer,
				Timestamp: 200,
			},
			Checkpoint: ShastaCheckpoint{
				BlockNumber: 11,
				BlockHash:   common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ee"),
				StateRoot:   common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ff"),
			},
		},
	}

	ordered := []ProofCarryData{pcd1, pcd2}
	assert.True(t, ValidateShastaProofCarryDataVec(ordered), "ordered proof carry data should validate")

	unordered := []ProofCarryData{pcd2, pcd1}
	assert.False(t, ValidateShastaProofCarryDataVec(unordered), "unordered proof carry data should fail validation")

	_, err := ShastaPCDAggregationHash(unordered, common.Address{})
	assert.Error(t, err, "unordered proof carry data should return error")
}
