package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type AccessList []*AccessTuple

func (a AccessList) GethType() types.AccessList {
	if a == nil {
		log.Warn("missing AccessList when converting to GethType")
		return nil
	}
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
