package witness

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/gaiko/pkg/mpt"
)

func (g *GuestInput) NewWitness() (*stateless.Witness, error) {
	wit := new(stateless.Witness)
	// set headers
	wit.Headers = append([]*types.Header{g.ParentHeader}, g.AncestorHeaders...)
	wit.State = map[string]struct{}{}
	wit.Codes = map[string]struct{}{}
	contracts := make(map[common.Hash][]byte, len(g.Contracts))
	for _, contract := range g.Contracts {
		codeHash := keccak.Keccak(contract)
		contracts[codeHash] = contract
	}
	rootBytes, err := rlp.EncodeToBytes(g.ParentStateTrie)
	if err != nil {
		return nil, err
	}
	wit.State[string(rootBytes)] = struct{}{}
	for addr, storage := range g.ParentStorage {
		acc, accBytes, err := getAccount(g.ParentStateTrie, addr)
		if err != nil {
			if err == ErrNotFound {
				log.Warn("account not found", "address", addr)
				acc = types.NewEmptyStateAccount()
			} else {
				return nil, err
			}
		}
		// set accounts
		wit.State[string(accBytes)] = struct{}{}
		root, err := storage.Trie.Hash()
		if err != nil {
			return nil, err
		}
		if root != acc.Root {
			return nil, fmt.Errorf("account root mismatch for address: %#x", addr)
		}

		var code []byte
		if common.BytesToHash(acc.CodeHash) != types.EmptyCodeHash {
			code = contracts[common.BytesToHash(acc.CodeHash)]
			if code == nil {
				return nil, errors.New("missing code")
			}
		}
		// set codes
		wit.Codes[string(code)] = struct{}{}
		for _, slot := range storage.Slots {
			key := common.BigToHash(slot)
			value, err := getStorage(storage.Trie, key)
			if err != nil {
				if err != ErrNotFound {
					return nil, err
				}
				log.Warn("slot not found", "key", key)
			}
			wit.State[string(value)] = struct{}{}
		}
	}

	return wit, nil
}

var ErrNotFound = errors.New("not found")

func getAccount(trie *mpt.MptNode, address common.Address) (*types.StateAccount, []byte, error) {
	accBytes, err := trie.Get(keccak.Keccak(address.Bytes()).Bytes())
	if err != nil {
		return nil, nil, err
	}
	if accBytes == nil {
		return nil, nil, ErrNotFound
	}

	acc := new(types.StateAccount)
	err = rlp.DecodeBytes(accBytes, acc)
	if err != nil {
		return nil, nil, err
	}
	return acc, accBytes, nil
}

func getStorage(trie *mpt.MptNode, key common.Hash) ([]byte, error) {
	return trie.Get(keccak.Keccak(key.Bytes()).Bytes())
}
