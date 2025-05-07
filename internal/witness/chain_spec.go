package witness

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
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
		if len(chainSpec.HardForks) != len(other.HardForks) {
			return errors.New("unexpected hard_forks")
		}
		for idx, fork := range chainSpec.HardForks {
			if fork.SpecID != other.HardForks[idx].SpecID {
				return errors.New("unexpected hard_forks")
			}
			if fork.Condition != other.HardForks[idx].Condition {
				return errors.New("unexpected hard_forks")
			}
		}
		if !chainSpec.Eip1559Constants.Equal(other.Eip1559Constants) {
			return errors.New("unexpected eip_1559_constants")
		}
		if !cmpAddress(chainSpec.L1Contract, other.L1Contract) {
			return errors.New("unexpected l1_contract")
		}

		if !cmpAddress(chainSpec.L2Contract, other.L2Contract) {
			return errors.New("unexpected l2_contract")
		}
		if chainSpec.IsTaiko != other.IsTaiko {
			return errors.New("unexpected is_taiko")
		}
	}
	return nil
}

func cmpAddress(a, b *common.Address) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Cmp(*b) == 0
}

type SpecID string
type ProofType string

const (
	NativeProofType  ProofType = "NATIVE"
	Sp1ProofType     ProofType = "SP1"
	SGXProofType     ProofType = "SGX"
	Risc0ProofType   ProofType = "RISC0"
	SGXGethProofType ProofType = "SGXGETH"
)

type HardFork struct {
	SpecID    SpecID
	Condition ForkCondition
}

type HardForks []*HardFork

//	{
//		"HEKLA": {
//			"Block": 0
//		},
//		"ONTAKE": {
//			"Block": 0
//		},
//		"PACAYA": {
//			"Block": 10
//		},
//		"CANCUN": "TBD"
//	}
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
				pairInner, ok := iter()
				if !ok {
					break
				}
				key := pairInner.Key
				value := pairInner.Value
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

type VerifierAddressFork map[ProofType]*common.Address

func (vf *VerifierAddressFork) UnmarshalJSON(data []byte) error {
	*vf = make(VerifierAddressFork)
	var m map[string]*common.Address
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for k, v := range m {
		(*vf)[ProofType(strings.ToUpper(k))] = v
	}
	return nil
}

type Network string

const (
	TaikoA7Network      Network = "taiko_a7"
	TaikoMainnetNetwork Network = "taiko_mainnet"
	EthereumNetwork     Network = "ethereum"
	HoleskyNetwork      Network = "holesky"
	TaikoDevNetwork     Network = "taiko_dev"
)

//go:generate go run github.com/fjl/gencodec -type ChainSpec -out gen_chain_spec.go
type ChainSpec struct {
	Name                 Network                        `json:"name"                   gencodec:"required"`
	ChainID              uint64                         `json:"chain_id"               gencodec:"required"`
	MaxSpecID            SpecID                         `json:"max_spec_id"            gencodec:"required"`
	HardForks            HardForks                      `json:"hard_forks"             gencodec:"required"`
	Eip1559Constants     *Eip1559Constants              `json:"eip_1559_constants"     gencodec:"required"`
	L1Contract           *common.Address                `json:"l1_contract"`
	L2Contract           *common.Address                `json:"l2_contract"`
	RPC                  string                         `json:"rpc"                    gencodec:"required"`
	BeaconRPC            *string                        `json:"beacon_rpc"`
	VerifierAddressForks map[SpecID]VerifierAddressFork `json:"verifier_address_forks" gencodec:"required"`
	GenesisTime          uint64                         `json:"genesis_time"           gencodec:"required"`
	SecondsPerSlot       uint64                         `json:"seconds_per_slot"       gencodec:"required"`
	IsTaiko              bool                           `json:"is_taiko"               gencodec:"required"`
}

var _ json.Unmarshaler = (*ChainSpec)(nil)

func (c *ChainSpec) getForkVerifierAddress(
	blockNum uint64,
	proofType ProofType,
) common.Address {
	for _, fork := range slices.Backward(c.HardForks) {
		if fork.Condition.Active(blockNum, 0) {
			if verifierAddressFork, ok := c.VerifierAddressForks[fork.SpecID]; ok {
				verifierAddress := verifierAddressFork[proofType]
				if verifierAddress == nil {
					return common.Address{}
				}
				return *verifierAddress
			}
		}
	}
	return common.Address{}
}

func (c *ChainSpec) chainConfig() (*params.ChainConfig, error) {
	switch c.Name {
	case TaikoA7Network:
		chainConfig := params.NetworkIDToChainConfigOrDefault(params.HeklaNetworkID)
		chainConfig.ChainID = params.HeklaNetworkID
		chainConfig.OntakeBlock = core.HeklaOntakeBlock
		chainConfig.PacayaBlock = core.HeklaPacayaBlock
		return chainConfig, nil
	case TaikoMainnetNetwork:
		chainConfig := params.NetworkIDToChainConfigOrDefault(params.TaikoMainnetNetworkID)
		chainConfig.ChainID = params.TaikoMainnetNetworkID
		chainConfig.OntakeBlock = core.MainnetOntakeBlock
		chainConfig.PacayaBlock = core.MainnetPacayaBlock
		return chainConfig, nil
	case EthereumNetwork:
		return params.MainnetChainConfig, nil
	case HoleskyNetwork:
		return params.HoleskyChainConfig, nil
	case TaikoDevNetwork:
		chainConfig := params.NetworkIDToChainConfigOrDefault(params.TaikoInternalL2ANetworkID)
		chainConfig.ChainID = params.TaikoInternalL2ANetworkID
		chainConfig.OntakeBlock = core.InternalDevnetOntakeBlock
		chainConfig.PacayaBlock = core.InternalDevnetPacayaBlock
		return chainConfig, nil
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

func (e *Eip1559Constants) Equal(other *Eip1559Constants) bool {
	if e == nil || other == nil {
		return true
	}
	return e.BaseFeeChangeDenominator.Cmp(other.BaseFeeChangeDenominator) == 0 &&
		e.BaseFeeMaxIncreaseDenominator.Cmp(other.BaseFeeMaxIncreaseDenominator) == 0 &&
		e.BaseFeeMaxDecreaseDenominator.Cmp(other.BaseFeeMaxDecreaseDenominator) == 0 &&
		e.ElasticityMultiplier.Cmp(other.ElasticityMultiplier) == 0
}

type eip1559ConstantsMarshaling struct {
	BaseFeeChangeDenominator      *hexutil.Big `json:"base_fee_change_denominator"       gencodec:"required"`
	BaseFeeMaxIncreaseDenominator *hexutil.Big `json:"base_fee_max_increase_denominator" gencodec:"required"`
	BaseFeeMaxDecreaseDenominator *hexutil.Big `json:"base_fee_max_decrease_denominator" gencodec:"required"`
	ElasticityMultiplier          *hexutil.Big `json:"elasticity_multiplier"             gencodec:"required"`
}
