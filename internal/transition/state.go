package transition

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/taikoxyz/gaiko/internal"
)

type preState struct {
	statedb  *state.StateDB
	getHash  func(num uint64) common.Hash
	accounts map[common.Address]*types.StateAccount
}

// makePreStateAndGetHash initializes the state database with the provided guest input data,
// commits the state, and returns the state database along with a function to retrieve historical
// block hashes.
//
// The function performs the following steps:
// 1. Creates an in-memory database and a trie database.
// 2. Initializes a new state database and state DB instance.
// 3. Processes the contracts and parent storage to set up the state DB.
// 4. Commits the state and reinitializes the state DB with the new root hash.
// 5. Constructs a map of historical block hashes.
//
// Returns:
// - *state.StateDB: The initialized state database.
// - func(num uint64) common.Hash: A function to retrieve historical block hashes by block number.
// - error: An error if any occurs during the process.
//
// *Note*:
// This StateDB is only used for execution without trust its root.
func (g *GuestInput) makePreState() (*preState, error) {
	parentRoot, err := g.ParentStateTrie.Hash()
	if err != nil {
		return nil, err
	}
	if g.ParentHeader.Root != parentRoot {
		return nil, fmt.Errorf("parent state root mismatch: expected %s, got %s",
			g.ParentHeader.Root.Hex(), parentRoot.Hex())
	}
	mdb := rawdb.NewMemoryDatabase()
	tdb := triedb.NewDatabase(mdb, &triedb.Config{Preimages: true})
	sdb := state.NewDatabase(tdb, nil)
	statedb, _ := state.New(types.EmptyRootHash, sdb)
	contracts := make(map[common.Hash][]byte, len(g.Contracts))
	for _, contract := range g.Contracts {
		codeHash := internal.Keccak(contract)
		contracts[common.BytesToHash(codeHash)] = contract
	}
	accounts := make(map[common.Address]*types.StateAccount, len(g.ParentStorage))
	for addr, storage := range g.ParentStorage {
		acc, err := getAccount(g.ParentStateTrie, addr)
		if err != nil {
			if err == ErrNotFound {
				acc = types.NewEmptyStateAccount()
			} else {
				return nil, err
			}
		} else {
			// skip not found accounts
			accounts[addr] = acc
		}
		root, err := storage.Trie.Hash()
		if err != nil {
			return nil, err
		}
		if root != acc.Root {
			return nil, fmt.Errorf("account root mismatch for address: %s", addr.Hex())
		}
		var code []byte
		if common.BytesToHash(acc.CodeHash) != types.EmptyCodeHash {
			code = contracts[common.BytesToHash(acc.CodeHash)]
			if code == nil {
				return nil, errors.New("missing code")
			}
		}
		statedb.SetCode(addr, code)
		statedb.SetNonce(addr, acc.Nonce)
		statedb.SetBalance(addr, acc.Balance, tracing.BalanceIncreaseGenesisBalance)
		for _, slot := range storage.Slots {
			key := common.BigToHash(slot)
			value, err := getStorage(g.ParentStateTrie, key)
			if err != nil && err != ErrNotFound {
				return nil, err
			}
			statedb.SetState(addr, key, value)
		}
	}
	root, _ := statedb.Commit(0, false)
	statedb, _ = state.New(root, sdb)

	historyHashes := make(map[uint64]common.Hash, len(g.AncestorHeaders)+1)
	historyHashes[g.ParentHeader.Number.Uint64()] = g.ParentHeader.Hash()
	prev := g.ParentHeader
	for _, header := range g.AncestorHeaders {
		if prev.ParentHash != header.Hash() {
			return nil, fmt.Errorf(
				"parent hash mismatch: expected %s, got %s",
				prev.ParentHash.Hex(),
				header.Hash().Hex(),
			)
		}
		historyHashes[header.Number.Uint64()] = header.Hash()
		prev = header
	}
	return &preState{
		statedb: statedb,
		getHash: func(num uint64) common.Hash {
			return historyHashes[num]
		},
		accounts: accounts,
	}, nil
}
