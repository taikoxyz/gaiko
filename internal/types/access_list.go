package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type AccessList []*AccessTuple

func (a AccessList) GethType() types.AccessList {
	accessList := make(types.AccessList, len(a))
	for i, accessTuple := range a {
		accessList[i] = types.AccessTuple(*accessTuple)
	}
	return accessList
}

type AccessTuple struct {
	Address     common.Address `json:"address"      gencodec:"required"`
	StorageKeys []common.Hash  `json:"storage_keys" gencodec:"required"`
}
