package t8n

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

var ErrNotFound = errors.New("not found")

func getAccount(trie *trie.Trie, address common.Address) (*types.StateAccount, error) {
	res, err := trie.Get(keccak(address.Bytes()))
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

func getStorage(trie *trie.Trie, key common.Hash) (common.Hash, error) {
	enc, err := trie.Get(keccak(key.Bytes()))
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

func updateAccount(trie *trie.Trie, address common.Address, acc *types.StateAccount) error {
	hk := keccak(address.Bytes())
	data, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return err
	}
	return trie.Update(hk, data)
}

func updateStorage(trie *trie.Trie, key, value []byte) error {
	hk := keccak(key)
	v, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}
	return trie.Update(hk, v)
}
