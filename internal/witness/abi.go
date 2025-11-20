package witness

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/encoding"
)

var (
	stringTy, _          = abi.NewType("string", "", nil)
	uint64Ty, _          = abi.NewType("uint64", "", nil)
	addressTy, _         = abi.NewType("address", "", nil)
	byte32Ty, _          = abi.NewType("bytes32", "", nil)
	byte32sTy, _         = abi.NewType("bytes32[]", "", nil)
	transitionComponents = []abi.ArgumentMarshaling{
		{
			Name: "parentHash",
			Type: "bytes32",
		},
		{
			Name: "blockHash",
			Type: "bytes32",
		},
		{
			Name: "stateRoot",
			Type: "bytes32",
		},
		{
			Name: "graffiti",
			Type: "bytes32",
		},
	}
	ontakeTransitionComponentsType, _ = abi.NewType("tuple", "TaikoData.Transition", transitionComponents)
	pacayaTransitionComponentsType, _ = abi.NewType(
		"tuple",
		"ITaikoInbox.Transition",
		encoding.BatchTransitionComponents,
	)
	publicInputsV1Type = abi.Arguments{
		{Name: "VERIFY_PROOF", Type: stringTy},
		{Name: "_chainId", Type: uint64Ty},
		{Name: "_verifierContract", Type: addressTy},
		{Name: "_transition", Type: ontakeTransitionComponentsType},
		{Name: "_newInstance", Type: addressTy},
		{Name: "_prover", Type: addressTy},
		{Name: "_metaHash", Type: byte32Ty},
	}
	publicInputsV2Type = abi.Arguments{
		{Name: "VERIFY_PROOF", Type: stringTy},
		{Name: "_chainId", Type: uint64Ty},
		{Name: "_verifierContract", Type: addressTy},
		{Name: "_transition", Type: pacayaTransitionComponentsType},
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

func init() {
	arg, err := findArgumentInEventInputs(batchProposedEvent.Inputs, "meta")
	if err != nil {
		panic(err)
	}
	batchMetadataComponentsArgs = abi.Arguments{arg}
	arg, err = findArgumentInEventInputs(batchProposedEvent.Inputs, "info")
	if err != nil {
		panic(err)
	}
	batchInfoComponentsArgs = abi.Arguments{arg}
	arg, err = findArgumentInEventInputs(blockProposedEvent.Inputs, "meta")
	if err != nil {
		panic(err)
	}
	blockMetadataComponentsArgs = abi.Arguments{arg}
	arg, err = findArgumentInEventInputs(blockProposedV2Event.Inputs, "meta")
	if err != nil {
		panic(err)
	}
	blockMetadataV2ComponentsArgs = abi.Arguments{arg}
}

// ABIEncoder is an interface for solidity structs encoding
// See [`abi.encode`](https://docs.soliditylang.org/en/latest/abi-spec.html)
type ABIEncoder interface {
	ABIEncode() ([]byte, error)
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

const signalSlots = "_signalSlots"

// decode `_signalSlots` from `anchorV3` transaction
/*
function anchorV3(
	uint64 _anchorBlockId,
	bytes32 _anchorStateRoot,
	uint32 _parentGasUsed,
	LibSharedData.BaseFeeConfig calldata _baseFeeConfig,
	bytes32[] calldata _signalSlots
)
*/
func decodeAnchorV3ArgsSignalSlots(input []byte) ([][32]byte, error) {
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
