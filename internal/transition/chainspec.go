package transition

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SpecId = uint8
type ProofType = uint8

const (
	SgxProofType ProofType = 2
)

type ChainSpec struct {
	Name                 string
	ChainId              uint64
	MaxSpecId            SpecId
	HardForks            map[SpecId]ForkCondition
	Eip1559Constants     Eip1559Constants
	L1Contract           *common.Address
	L2Contract           *common.Address
	RPC                  string
	BeaconRPC            *string
	VerifierAddressForks map[SpecId]map[ProofType]*common.Address
	GenesisTime          uint64
	SecondsPerSlot       uint64
	IsTaiko              bool
}

// pub fn get_fork_verifier_address(
// 	&self,
// 	block_num: u64,
// 	proof_type: ProofType,
// ) -> Result<Address> {
// 	// fall down to the first fork that is active as default
// 	for (spec_id, fork) in self.hard_forks.iter().rev() {
// 		if fork.active(block_num, 0u64) {
// 			if let Some(fork_verifier) = self.verifier_address_forks.get(spec_id) {
// 				return fork_verifier
// 					.get(&proof_type)
// 					.ok_or_else(|| anyhow!("Verifier type not found"))
// 					.and_then(|address| {
// 						address.ok_or_else(|| anyhow!("Verifier address not found"))
// 					});
// 			}
// 		}
// 	}

// 	Err(anyhow!("fork verifier is not active"))
// }

func (c *ChainSpec) getForkVerifierAddress(blockNum uint64) (common.Address, error) {
	// TODO: Implement this function
	return common.Address{}, nil
}

type ForkCondition interface {
	Active(blockNumber uint64, timestamp uint64) bool
}

type BlockNumber uint64

func (b BlockNumber) Active(blockNumber uint64, _ uint64) bool {
	return blockNumber >= uint64(b)
}

type BlockTimestamp uint64

func (b BlockTimestamp) Active(_ uint64, timestamp uint64) bool {
	return timestamp >= uint64(b)
}

type Eip1559Constants struct {
	BaseFeeChangeDenominator      *big.Int
	BaseFeeMaxIncreaseDenominator *big.Int
	BaseFeeMaxDecreaseDenominator *big.Int
	ElasticityMultiplier          *big.Int
}
