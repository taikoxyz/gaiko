package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type AccessList []AccessTuple

func (a AccessList) Origin() types.AccessList {
	accessList := make(types.AccessList, len(a))
	for i, accessTuple := range a {
		accessList[i] = accessTuple.Origin()
	}
	return accessList
}

type AccessTuple struct {
	Address     common.Address `json:"address" gencodec:"required"`
	StorageKeys []common.Hash  `json:"storage_keys" gencodec:"required"`
}

func (a *AccessTuple) Origin() types.AccessTuple {
	return types.AccessTuple{
		Address:     a.Address,
		StorageKeys: a.StorageKeys,
	}
}
