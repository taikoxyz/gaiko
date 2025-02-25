package witness

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
	byte32sTy, _     = abi.NewType("bytes32[]", "", nil)
	publicInputsType = abi.Arguments{
		{Name: "VERIFY_PROOF", Type: stringTy},
		{Name: "_chainId", Type: uint64Ty},
		{Name: "_verifierContract", Type: addressTy},
		{Name: "_transition", Type: encoding.TransitionComponentsType},
		{Name: "_newInstance", Type: addressTy},
		{Name: "_metaHash", Type: byte32Ty},
	}
	batchTxHashArgs = abi.Arguments{
		{Name: "_txListHash", Type: byte32Ty},
		{Name: "blobHashes_", Type: byte32sTy},
	}
	batchMetadataComponentsArgs   abi.Arguments
	batchInfoComponentsArgs       abi.Arguments
	blockMetadataComponentsArgs   abi.Arguments
	blockMetadataV2ComponentsArgs abi.Arguments
	batchProposedEvent            = encoding.TaikoInboxABI.Events["BatchProposed"]
	blockProposedEvent            = encoding.TaikoL1ABI.Events["BlockProposed"]
	blockProposedV2Event          = encoding.TaikoL1ABI.Events["BlockProposedV2"]
	anchorV3Method                = encoding.TaikoAnchorABI.Methods["anchorV3"]
)

const signalSlots = "_signalSlots"

func init() {
	arg, err := findArgumentInEventInputs(batchProposedEvent.Inputs, "meta")
	if err != nil {
		log.Crit("Get BatchProposed failed", err)
	}
	batchMetadataComponentsArgs = abi.Arguments{arg}
	arg, err = findArgumentInEventInputs(batchProposedEvent.Inputs, "info")
	if err != nil {
		log.Crit("Get BatchInfo failed", err)
	}
	batchInfoComponentsArgs = abi.Arguments{arg}
	arg, err = findArgumentInEventInputs(blockProposedEvent.Inputs, "meta")
	if err != nil {
		log.Crit("Get BlockProposed failed", err)
	}
	blockMetadataComponentsArgs = abi.Arguments{arg}
	arg, err = findArgumentInEventInputs(blockProposedV2Event.Inputs, "meta")
	if err != nil {
		log.Crit("Get BlockProposedV2 failed", err)
	}
	blockMetadataV2ComponentsArgs = abi.Arguments{arg}
}

// ABIEncoder is an interface for solidity structs encoding
// See [`abi.encode`](https://docs.soliditylang.org/en/latest/abi-spec.html)
type ABIEncoder interface {
	Encode() ([]byte, error)
}

// generated binding doesn't have any struct specs, we can find them in the used places
func findArgumentInEventInputs(inputs abi.Arguments, name string) (abi.Argument, error) {
	for _, input := range inputs {
		if input.Name == name {
			return input, nil
		}
	}
	return abi.Argument{}, errors.New("input not found")
}

func decodeAnchorV3Args_signalSlots(input []byte) ([][32]byte, error) {
	args := map[string]any{}
	err := anchorV3Method.Inputs.UnpackIntoMap(args, input)
	if err != nil {
		return nil, err
	}
	signalSlots, ok := args[signalSlots].([][32]byte)
	if !ok {
		return nil, errors.New("signalSlots not found")
	}
	return signalSlots, nil
}
