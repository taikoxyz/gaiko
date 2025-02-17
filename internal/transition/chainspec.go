package transition

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type SpecID = uint8
type ProofType = uint8

const (
	NativeProofType   ProofType = 0
	Sp1ProofType      ProofType = 1
	SgxProofType      ProofType = 2
	Risc0ProofType    ProofType = 3
	GaikoSgxProofType ProofType = 4
)

type VerifierAddressFork struct {
	SpecId    SpecID
	Verifiers map[ProofType]*common.Address
}

type HardFork struct {
	SpecID    SpecID
	Condition ForkCondition
}

type ChainSpec struct {
	Name                 string
	ChainID              uint64
	MaxSpecID            SpecID
	HardForks            []HardFork
	Eip1559Constants     Eip1559Constants
	L1Contract           *common.Address
	L2Contract           *common.Address
	RPC                  string
	BeaconRPC            *string
	VerifierAddressForks map[SpecID]map[ProofType]*common.Address
	GenesisTime          uint64
	SecondsPerSlot       uint64
	IsTaiko              bool
}

func (c *ChainSpec) getForkVerifierAddress(blockNum uint64, proofType ProofType) (common.Address, error) {
	for i := len(c.HardForks) - 1; i >= 0; i-- {
		fork := c.HardForks[i]
		if fork.Condition.Active(blockNum, 0) {
			if verifierAddressFork, ok := c.VerifierAddressForks[fork.SpecID]; ok {
				verifierAddress := verifierAddressFork[proofType]
				if verifierAddress == nil {
					return common.Address{}, fmt.Errorf("fork verifier for proof type %d is not active", proofType)
				}
				return *verifierAddress, nil
			}
		}
	}
	return common.Address{}, fmt.Errorf("fork verifier is not active")
}

func (c *ChainSpec) chainConfig() (*params.ChainConfig, error) {
	switch c.Name {
	case "taiko_a7":
		return params.NetworkIDToChainConfigOrDefault(params.HeklaNetworkID), nil
	case "taiko_mainnet":
		return params.NetworkIDToChainConfigOrDefault(params.TaikoMainnetNetworkID), nil
	case "ethereum":
		return params.MainnetChainConfig, nil
	case "holesky":
		return params.HoleskyChainConfig, nil
	case "taiko_dev":
		return params.NetworkIDToChainConfigOrDefault(params.TaikoInternalL2ANetworkID), nil
	default:
		return nil, errors.New("unsupported chain spec")
	}
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
