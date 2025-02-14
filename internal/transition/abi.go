package transition

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/encoding"
)

var (
	stringTy, _             = abi.NewType("string", "", nil)
	uint64Ty, _             = abi.NewType("uint64", "", nil)
	addressTy, _            = abi.NewType("address", "", nil)
	byte32Ty, _             = abi.NewType("bytes32", "", nil)
	blockMetadataComponents = []abi.ArgumentMarshaling{
		{
			Name: "l1Hash",
			Type: "bytes32",
		},
		{
			Name: "difficulty",
			Type: "bytes32",
		},
		{
			Name: "blobHash",
			Type: "bytes32",
		},
		{
			Name: "extraData",
			Type: "bytes32",
		},
		{
			Name: "depositsHash",
			Type: "bytes32",
		},
		{
			Name: "coinbase",
			Type: "address",
		},
		{
			Name: "id",
			Type: "uint64",
		},
		{
			Name: "gasLimit",
			Type: "uint32",
		},
		{
			Name: "timestamp",
			Type: "uint64",
		},
		{
			Name: "l1Height",
			Type: "uint64",
		},
		{
			Name: "minTier",
			Type: "uint16",
		},
		{
			Name: "blobUsed",
			Type: "bool",
		},
		{
			Name: "parentMetaHash",
			Type: "bytes32",
		},
		{
			Name: "sender",
			Type: "address",
		},
	}

	publicInputsType = abi.Arguments{
		{Name: "VERIFY_PROOF", Type: stringTy},
		{Name: "_chainId", Type: uint64Ty},
		{Name: "_verifierContract", Type: addressTy},
		{Name: "_transition", Type: encoding.TransitionComponentsType},
		{Name: "_newInstance", Type: addressTy},
		{Name: "_metaHash", Type: byte32Ty},
	}
	blockMetadataComponentsType, _      = abi.NewType("tuple", "TaikoData.BlockMetadata", blockMetadataComponents)
	blockMetadataComponentsArgs         = abi.Arguments{{Name: "TaikoData.BlockMetadata", Type: blockMetadataComponentsType}}
	blockMetadataV2ComponentsArgs       = abi.Arguments{{Name: "TaikoData.BlockMetadataV2", Type: encoding.BlockMetadataV2ComponentsType}}
	batchMetaDataComponentsArrayType, _ = abi.NewType("tuple", "ITaikoInbox.BatchMetadata", encoding.BatchMetaDataComponents)
)

type AbiEncoder interface {
	Encode() ([]byte, error)
}
