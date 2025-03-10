package transition

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/gaiko/pkg/mpt"
)

var ErrNotFound = errors.New("not found")

func getAccount(trie *mpt.MptNode, address common.Address) (*types.StateAccount, error) {
	res, err := trie.Get(keccak.Keccak(address.Bytes()).Bytes())
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}

	ret := new(types.StateAccount)
	err = rlp.DecodeBytes(res, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func getStorage(trie *mpt.MptNode, key common.Hash) (common.Hash, error) {
	enc, err := trie.Get(keccak.Keccak(key.Bytes()).Bytes())
	if err != nil {
		return common.Hash{}, err
	}
	if len(enc) == 0 {
		return common.Hash{}, ErrNotFound
	}

	_, content, _, err := rlp.Split(enc)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(content), nil
}

func updateAccount(trie *mpt.MptNode, address common.Address, acc *types.StateAccount) error {
	hk := keccak.Keccak(address.Bytes())
	data, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return err
	}
	_, err = trie.Insert(hk.Bytes(), data)
	return err
}

func updateStorage(trie *mpt.MptNode, key, value []byte) error {
	hk := keccak.Keccak(key)
	v, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}
	_, err = trie.Insert(hk.Bytes(), v)
	return err
}
