package transition

import (
	"errors"
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/taikoxyz/gaiko/pkg/keccak"
)

type preState struct {
	stateDB  *state.StateDB
	getHash  func(num uint64) common.Hash
	accounts map[common.Address]*types.StateAccount
}

// newPreState initializes the state database with the provided guest input data,
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
func newPreState(g *witness.GuestInput) (*preState, error) {
	parentRoot, err := g.ParentStateTrie.Hash()
	if err != nil {
		return nil, err
	}
	if g.ParentHeader.Root != parentRoot {
		return nil, fmt.Errorf("parent state root mismatch: expected %#x, got %#x",
			g.ParentHeader.Root, parentRoot)
	}
	mdb := rawdb.NewMemoryDatabase()
	tdb := triedb.NewDatabase(mdb, triedb.HashDefaults)
	sdb := state.NewDatabase(tdb, nil)
	stateDB, err := state.New(types.EmptyRootHash, sdb)
	if err != nil {
		return nil, err
	}
	contracts := make(map[common.Hash][]byte, len(g.Contracts))
	for _, contract := range g.Contracts {
		codeHash := keccak.Keccak(contract)
		contracts[codeHash] = contract
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
			return nil, fmt.Errorf("account root mismatch for address: %#x", addr)
		}
		var code []byte
		if common.BytesToHash(acc.CodeHash) != types.EmptyCodeHash {
			code = contracts[common.BytesToHash(acc.CodeHash)]
			if code == nil {
				return nil, errors.New("missing code")
			}
		}
		stateDB.SetCode(addr, code)
		stateDB.SetNonce(addr, acc.Nonce, tracing.NonceChangeGenesis)
		stateDB.SetBalance(addr, acc.Balance, tracing.BalanceIncreaseGenesisBalance)
		for _, slot := range storage.Slots {
			key := common.BigToHash(slot)
			value, err := getStorage(storage.Trie, key)
			if err != nil && err != ErrNotFound {
				return nil, err
			}
			stateDB.SetState(addr, key, value)
		}
	}
	root, err := stateDB.Commit(0, false, false)
	if err != nil {
		return nil, err
	}
	stateDB, err = state.New(root, sdb)
	if err != nil {
		return nil, err
	}

	historyHashes := make(map[uint64]common.Hash, len(g.AncestorHeaders)+1)
	historyHashes[g.ParentHeader.Number.Uint64()] = g.ParentHeader.Hash()
	prev := g.ParentHeader
	for header := range slices.Values(g.AncestorHeaders) {
		if prev.ParentHash != header.Hash() {
			return nil, fmt.Errorf(
				"parent hash mismatch: expected %#x, got %#x",
				prev.ParentHash,
				header.Hash(),
			)
		}
		historyHashes[header.Number.Uint64()] = header.Hash()
		prev = header
	}
	return &preState{
		stateDB: stateDB,
		getHash: func(num uint64) common.Hash {
			return historyHashes[num]
		},
		accounts: accounts,
	}, nil
}
