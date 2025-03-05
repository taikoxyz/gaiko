package witness

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"gitlab.com/c0b/go-ordered-json"
)

//go:embed chain_spec_list_default.json
var supportedChainSpecsJSON []byte

type SupportedChainSpecs []*ChainSpec

var defaultSupportedChainSpecs SupportedChainSpecs

func init() {
	if err := json.Unmarshal(supportedChainSpecsJSON, &defaultSupportedChainSpecs); err != nil {
		panic(err)
	}
}

func (s SupportedChainSpecs) verifyChainSpec(other *ChainSpec) error {
	for chainSpec := range slices.Values(s) {
		if chainSpec.ChainID != other.ChainID {
			continue
		}
		if chainSpec.MaxSpecID != other.MaxSpecID {
			return errors.New("unexpected max_spec_id")
		}
		if !slices.Equal(chainSpec.HardForks, other.HardForks) {
			return errors.New("unexpected hard_forks")
		}
		if chainSpec.Eip1559Constants != other.Eip1559Constants {
			return errors.New("unexpected eip_1559_constants")
		}
		if chainSpec.L1Contract != other.L1Contract {
			return errors.New("unexpected l1_contract")
		}
		if chainSpec.L2Contract != other.L2Contract {
			return errors.New("unexpected l2_contract")
		}
		if chainSpec.IsTaiko != other.IsTaiko {
			return errors.New("unexpected is_taiko")
		}
	}
	return nil
}

type SpecID string
type ProofType string

const (
	NativeProofType ProofType = "NATIVE"
	Sp1ProofType    ProofType = "SP1"
	SGXProofType    ProofType = "SGX"
	Risc0ProofType  ProofType = "RISC0"
)

type HardFork struct {
	SpecID    SpecID
	Condition ForkCondition
}

type HardForks []*HardFork

func (h *HardForks) UnmarshalJSON(data []byte) error {
	orderedMap := ordered.NewOrderedMap()
	if err := json.Unmarshal(data, orderedMap); err != nil {
		return err
	}
	iter := orderedMap.EntriesIter()
	for {
		pair, ok := iter()
		if !ok {
			break
		}
		switch val := pair.Value.(type) {
		case *ordered.OrderedMap:
			iter := val.EntriesIter()
			for {
				pair, ok := iter()
				if !ok {
					break
				}
				key := pair.Key
				value := pair.Value
				switch key {
				case "Block":
					valueNumber, err := value.(json.Number).Int64()
					if err != nil {
						return err
					}
					blockNumber := BlockNumber(uint64(valueNumber))
					*h = append(*h, &HardFork{
						SpecID:    SpecID(pair.Key),
						Condition: blockNumber,
					})
				case "Timestamp":
					valueNumber, err := value.(json.Number).Int64()
					if err != nil {
						return err
					}
					blockTimestamp := BlockTimestamp(uint64(valueNumber))
					*h = append(*h, &HardFork{
						SpecID:    SpecID(pair.Key),
						Condition: blockTimestamp,
					})
				default:
					return fmt.Errorf("unknown key %s", key)
				}
			}
		case string:
			if val != "TBD" {
				return fmt.Errorf("unsupported hardfork: %s", val)
			}
			*h = append(*h, &HardFork{
				SpecID:    SpecID(pair.Key),
				Condition: TBD{},
			})
		default:
			return fmt.Errorf("unsupported type for hardfork: %T", val)
		}
	}

	return nil
}

//go:generate go run github.com/fjl/gencodec -type ChainSpec -out gen_chain_spec.go
type ChainSpec struct {
	Name                 string                                   `json:"name"                   gencodec:"required"`
	ChainID              uint64                                   `json:"chain_id"               gencodec:"required"`
	MaxSpecID            SpecID                                   `json:"max_spec_id"            gencodec:"required"`
	HardForks            HardForks                                `json:"hard_forks"             gencodec:"required"`
	Eip1559Constants     Eip1559Constants                         `json:"eip_1559_constants"     gencodec:"required"`
	L1Contract           *common.Address                          `json:"l1_contract"`
	L2Contract           *common.Address                          `json:"l2_contract"`
	RPC                  string                                   `json:"rpc"                    gencodec:"required"`
	BeaconRPC            *string                                  `json:"beacon_rpc"`
	VerifierAddressForks map[SpecID]map[ProofType]*common.Address `json:"verifier_address_forks" gencodec:"required"`
	GenesisTime          uint64                                   `json:"genesis_time"           gencodec:"required"`
	SecondsPerSlot       uint64                                   `json:"seconds_per_slot"       gencodec:"required"`
	IsTaiko              bool                                     `json:"is_taiko"               gencodec:"required"`
}

var _ json.Unmarshaler = (*ChainSpec)(nil)

func (c *ChainSpec) getForkVerifierAddress(
	blockNum uint64,
	proofType ProofType,
) (common.Address, error) {
	for _, fork := range slices.Backward(c.HardForks) {
		if fork.Condition.Active(blockNum, 0) {
			if verifierAddressFork, ok := c.VerifierAddressForks[fork.SpecID]; ok {
				verifierAddress := verifierAddressFork[proofType]
				if verifierAddress == nil {
					return common.Address{}, fmt.Errorf(
						"fork verifier for proof type %s is not active",
						proofType,
					)
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

type TBD struct{}

func (t TBD) Active(_ uint64, _ uint64) bool {
	return false
}

//go:generate go run github.com/fjl/gencodec -type Eip1559Constants -field-override eip1559ConstantsMarshaling -out gen_eip1559_constants.go
type Eip1559Constants struct {
	BaseFeeChangeDenominator      *big.Int `json:"base_fee_change_denominator"       gencodec:"required"`
	BaseFeeMaxIncreaseDenominator *big.Int `json:"base_fee_max_increase_denominator" gencodec:"required"`
	BaseFeeMaxDecreaseDenominator *big.Int `json:"base_fee_max_decrease_denominator" gencodec:"required"`
	ElasticityMultiplier          *big.Int `json:"elasticity_multiplier"             gencodec:"required"`
}

type eip1559ConstantsMarshaling struct {
	BaseFeeChangeDenominator      *hexutil.Big `json:"base_fee_change_denominator"       gencodec:"required"`
	BaseFeeMaxIncreaseDenominator *hexutil.Big `json:"base_fee_max_increase_denominator" gencodec:"required"`
	BaseFeeMaxDecreaseDenominator *hexutil.Big `json:"base_fee_max_decrease_denominator" gencodec:"required"`
	ElasticityMultiplier          *hexutil.Big `json:"elasticity_multiplier"             gencodec:"required"`
}
