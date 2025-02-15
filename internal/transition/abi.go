package transition

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/encoding"
)

var (
	stringTy, _      = abi.NewType("string", "", nil)
	uint64Ty, _      = abi.NewType("uint64", "", nil)
	addressTy, _     = abi.NewType("address", "", nil)
	byte32Ty, _      = abi.NewType("bytes32", "", nil)
	publicInputsType = abi.Arguments{
		{Name: "VERIFY_PROOF", Type: stringTy},
		{Name: "_chainId", Type: uint64Ty},
		{Name: "_verifierContract", Type: addressTy},
		{Name: "_transition", Type: encoding.TransitionComponentsType},
		{Name: "_newInstance", Type: addressTy},
		{Name: "_metaHash", Type: byte32Ty},
	}
	batchMetaDataComponentsArgs   abi.Arguments
	blockMetadataComponentsArgs   abi.Arguments
	blockMetadataV2ComponentsArgs abi.Arguments
	batchProposedEvent            = encoding.TaikoInboxABI.Events["BatchProposed"]
	blockProposedEvent            = encoding.TaikoL1ABI.Events["BlockProposed"]
	blockProposedV2Event          = encoding.TaikoL1ABI.Events["BlockProposedV2"]
)

func init() {
	input, err := findInput("meta", batchProposedEvent.Inputs)
	if err != nil {
		log.Crit("Get BatchProposed failed", err)
	}
	batchMetaDataComponentsArgs = abi.Arguments{input}
	input, err = findInput("meta", blockProposedEvent.Inputs)
	if err != nil {
		log.Crit("Get BlockProposed failed", err)
	}
	blockMetadataComponentsArgs = abi.Arguments{input}
	input, err = findInput("meta", blockProposedV2Event.Inputs)
	if err != nil {
		log.Crit("Get BatchProposed failed", err)
	}
	blockMetadataV2ComponentsArgs = abi.Arguments{input}
}

type ABIEncoder interface {
	Encode() ([]byte, error)
}

func findInput(name string, inputs abi.Arguments) (abi.Argument, error) {
	for _, input := range inputs {
		if input.Name == name {
			return input, nil
		}
	}
	return abi.Argument{}, errors.New("input not found")
}
