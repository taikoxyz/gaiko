package transition

import "github.com/ethereum/go-ethereum/accounts/abi"

var (
	stringTy, _          = abi.NewType("string", "", nil)
	uint64Ty, _          = abi.NewType("uint64", "", nil)
	addressTy, _         = abi.NewType("address", "", nil)
	byte32Ty, _          = abi.NewType("bytes32", "", nil)
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
	blockMetadataV2Components = []abi.ArgumentMarshaling{
		{
			Name: "anchorBlockHash",
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
			Name: "anchorBlockId",
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
			Name: "proposer",
			Type: "address",
		},
		{
			Name: "livenessBond",
			Type: "uint96",
		},
		{
			Name: "proposedAt",
			Type: "uint64",
		},
		{
			Name: "proposedIn",
			Type: "uint64",
		},
		{
			Name: "blobTxListOffset",
			Type: "uint32",
		},
		{
			Name: "blobTxListLength",
			Type: "uint32",
		},
		{
			Name: "blobIndex",
			Type: "uint8",
		},
		{
			Name: "baseFeeConfig",
			Type: "tuple",
			Components: []abi.ArgumentMarshaling{
				{
					Name: "adjustmentQuotient",
					Type: "uint8",
				},
				{
					Name: "sharingPctg",
					Type: "uint8",
				},
				{
					Name: "gasIssuancePerSecond",
					Type: "uint32",
				},
				{
					Name: "minGasExcess",
					Type: "uint64",
				},
				{
					Name: "maxGasIssuancePerBlock",
					Type: "uint32",
				},
			},
		},
	}

	batchMetaDataComponents = []abi.ArgumentMarshaling{
		{
			Name: "infoHash",
			Type: "bytes32",
		},
		{
			Name: "proposer",
			Type: "address",
		},
		{
			Name: "batchId",
			Type: "uint64",
		},
		{
			Name: "proposedAt",
			Type: "uint64",
		},
	}
	transitionComponentsType, _ = abi.NewType("tuple", "TaikoData.Transition", transitionComponents)
	publicInputsType            = abi.Arguments{
		{Name: "VERIFY_PROOF", Type: stringTy},
		{Name: "_chainId", Type: uint64Ty},
		{Name: "_verifierContract", Type: addressTy},
		{Name: "_transition", Type: transitionComponentsType},
		{Name: "_newInstance", Type: addressTy},
		{Name: "_metaHash", Type: byte32Ty},
	}
	blockMetadataComponentsType, _      = abi.NewType("tuple", "TaikoData.BlockMetadata", blockMetadataComponents)
	blockMetadataV2ComponentsType, _    = abi.NewType("tuple", "TaikoData.BlockMetadataV2", blockMetadataV2Components)
	blockMetadataComponentsArgs         = abi.Arguments{{Name: "TaikoData.BlockMetadata", Type: blockMetadataComponentsType}}
	blockMetadataV2ComponentsArgs       = abi.Arguments{{Name: "TaikoData.BlockMetadataV2", Type: blockMetadataV2ComponentsType}}
	batchMetaDataComponentsArrayType, _ = abi.NewType("tuple", "ITaikoInbox.BatchMetadata", batchMetaDataComponents)
)
