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

func (g *SingleGuestInput) NewWitness() (*stateless.Witness, error) {
	wit := new(stateless.Witness)
	// set headers
	wit.Headers = append([]*types.Header{g.ParentHeader}, g.AncestorHeaders...)
	wit.State = map[string]struct{}{}
	wit.Codes = map[string]struct{}{}
	for _, contract := range g.Contracts {
		wit.Codes[string(contract)] = struct{}{}
	}
	onRLP := func(data []byte) {
		wit.State[string(data)] = struct{}{}
	}
	parentRoot, err := g.ParentStateTrie.Hash(onRLP)
	if err != nil {
		return nil, err
	}
	if g.ParentHeader.Root != parentRoot {
		return nil, fmt.Errorf("parent state root mismatch: expected %#x, got %#x",
			g.ParentHeader.Root, parentRoot)
	}

	for addr, storage := range g.ParentStorage {
		acc, err := getAccount(g.ParentStateTrie, addr)
		if err != nil {
			log.Warn("account not found", "id", g.ID(), "address", addr, "err", err)
			acc = types.NewEmptyStateAccount()
		}
		root, err := storage.Trie.Hash(onRLP)
		if err != nil {
			return nil, err
		}
		if root != acc.Root {
			return nil, fmt.Errorf("account root mismatch for address: %#x", addr)
		}
	}

	return wit, nil
}

var ErrNotFound = errors.New("not found")

func getAccount(trie *mpt.MptNode, address common.Address) (*types.StateAccount, error) {
	accBytes, err := trie.Get(keccak.Keccak(address.Bytes()).Bytes())
	if err != nil {
		return nil, err
	}
	if accBytes == nil {
		return nil, ErrNotFound
	}

	acc := new(types.StateAccount)
	err = rlp.DecodeBytes(accBytes, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
