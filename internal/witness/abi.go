package witness

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/encoding"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/shasta"
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
	shastaAnchorV4Method          abi.Method
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

	shastaAnchorABI, err := abi.JSON(strings.NewReader(shasta.ShastaAnchorABI))
	if err != nil {
		panic(err)
	}
	shastaAnchorV4Method = shastaAnchorABI.Methods["anchorV4"]
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

func decodeShastaAnchorCheckpoint(input []byte) (*ShastaCheckpoint, error) {
	// L2 Shasta anchor transaction (Anchor.anchorV4) encodes a single checkpoint struct:
	// (uint48 blockNumber, bytes32 blockHash, bytes32 stateRoot) -> 3 static 32-byte words.
	// This path avoids relying on the L1 ShastaAnchor ABI which includes dynamic fields.
	if len(input) == 96 {
		blockNumber := new(big.Int).SetBytes(input[:32])
		if blockNumber.BitLen() > 48 {
			return nil, fmt.Errorf("invalid shasta checkpoint block number: %#x", blockNumber)
		}
		return &ShastaCheckpoint{
			BlockNumber: blockNumber.Uint64(),
			BlockHash:   common.BytesToHash(input[32:64]),
			StateRoot:   common.BytesToHash(input[64:96]),
		}, nil
	}

	if shastaAnchorV4Method.Name == "" {
		return nil, errors.New("shasta anchor ABI not initialized")
	}
	args := map[string]any{}
	if err := shastaAnchorV4Method.Inputs.UnpackIntoMap(args, input); err != nil {
		return nil, err
	}

	var param any
	if value, ok := args["_blockParams"]; ok {
		param = value
	} else if value, ok := args["_checkpoint"]; ok {
		param = value
	} else {
		return nil, errors.New("shasta anchor params not found")
	}

	return shastaCheckpointFromTuple(param)
}

func shastaCheckpointFromTuple(value any) (*ShastaCheckpoint, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, errors.New("nil shasta anchor param")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("unexpected shasta anchor param type")
	}

	numberField := rv.FieldByName("AnchorBlockNumber")
	if !numberField.IsValid() {
		numberField = rv.FieldByName("BlockNumber")
	}
	hashField := rv.FieldByName("AnchorBlockHash")
	if !hashField.IsValid() {
		hashField = rv.FieldByName("BlockHash")
	}
	stateField := rv.FieldByName("AnchorStateRoot")
	if !stateField.IsValid() {
		stateField = rv.FieldByName("StateRoot")
	}

	if !numberField.IsValid() || !hashField.IsValid() || !stateField.IsValid() {
		return nil, errors.New("invalid shasta anchor param fields")
	}

	var blockNumber uint64
	switch v := numberField.Interface().(type) {
	case *big.Int:
		blockNumber = v.Uint64()
	case uint64:
		blockNumber = v
	case uint32:
		blockNumber = uint64(v)
	default:
		if numberField.Kind() == reflect.Uint64 {
			blockNumber = numberField.Uint()
		} else {
			return nil, errors.New("unsupported block number type")
		}
	}

	blockHash, err := toCommonHash(hashField)
	if err != nil {
		return nil, err
	}
	stateRoot, err := toCommonHash(stateField)
	if err != nil {
		return nil, err
	}

	return &ShastaCheckpoint{
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		StateRoot:   stateRoot,
	}, nil
}

func toCommonHash(value reflect.Value) (common.Hash, error) {
	if !value.IsValid() {
		return common.Hash{}, errors.New("invalid hash value")
	}
	switch v := value.Interface().(type) {
	case [32]byte:
		return common.BytesToHash(v[:]), nil
	case common.Hash:
		return v, nil
	default:
		if value.Kind() == reflect.Array && value.Len() == 32 && value.Type().Elem().Kind() == reflect.Uint8 {
			var out [32]byte
			for i := 0; i < 32; i++ {
				out[i] = byte(value.Index(i).Uint())
			}
			return common.BytesToHash(out[:]), nil
		}
		return common.Hash{}, errors.New("unsupported hash type")
	}
}
