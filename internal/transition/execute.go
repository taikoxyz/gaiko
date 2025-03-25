package transition

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/witness"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"golang.org/x/sync/errgroup"
)

// ExecuteAndVerify executes and verifies the given arguments using the provided witness.
// It retrieves the chain configuration from the witness and processes each guest input
// concurrently using an error group.
func ExecuteAndVerify(
	ctx context.Context,
	args *flags.Arguments,
	input witness.WitnessInput,
) error {
	chainConfig, err := input.ChainConfig()
	if err != nil {
		return err
	}
	eg, ctx := errgroup.WithContext(ctx)
	for pair := range input.GuestInputs() {
		pair := pair // https://go.dev/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			return executeWitness(ctx, pair, chainConfig)
		})
	}
	return eg.Wait()
}

func executeWitness(
	_ context.Context,
	pair *witness.Pair,
	chainConfig *params.ChainConfig,
) error {
	g := pair.Input
	txs := pair.Txs
	wit, err := g.NewWitness()
	if err != nil {
		return err
	}
	block := g.Block.WithBody(types.Body{
		Transactions: txs,
		Uncles:       g.Block.Uncles(),
		Withdrawals:  g.Block.Withdrawals(),
	})
	stateRoot, _, err := core.ExecuteStateless(chainConfig, vm.Config{}, block, wit)
	if err != nil {
		return err
	}
	expected := g.Block.Root()
	if expected != stateRoot {
		return fmt.Errorf(
			"block %d root mismatch: expected %#x, got %#x",
			g.Block.NumberU64(),
			expected,
			stateRoot,
		)
	}
	return nil
}

func executeAndVerify(
	_ context.Context,
	pair *witness.Pair,
	chainConfig *params.ChainConfig,
) error {
	g := pair.Input
	txs := pair.Txs
	preState, err := newPreState(g)
	if err != nil {
		return err
	}
	stateDB, err := apply(
		vm.Config{},
		preState.stateDB,
		g.Block,
		txs,
		preState.getHash,
		chainConfig,
	)
	if err != nil {
		return err
	}
	collector := make(Dumper)
	stateDB.DumpToCollector(collector, nil)
	for addr := range preState.accounts {
		_, ok := collector[addr]
		if !ok {
			// Account is deleted
			key := keccak.Keccak(addr.Bytes())
			if _, err := g.ParentStateTrie.Delete(key.Bytes()); err != nil {
				return err
			}
		}
	}

	for addr, acc := range collector {
		entry, ok := g.ParentStorage[addr]
		if !ok {
			return fmt.Errorf("account not found for address: %#x", addr)
		}
		_, ok = preState.accounts[addr]
		if !ok {
			// New Account
			entry.Trie.Clear()
		}
		for slot, value := range acc.Storage {
			key := keccak.Keccak(slot.Bytes())
			if value == (common.Hash{}) {
				if _, err := entry.Trie.Delete(key.Bytes()); err != nil {
					return err
				}
			} else {
				if err := updateStorage(entry.Trie, slot.Bytes(), value.Bytes()); err != nil {
					return err
				}
			}
		}
		root, err := entry.Trie.Hash()
		if err != nil {
			return err
		}
		stateAcc := &types.StateAccount{
			Nonce:    acc.Nonce,
			Balance:  new(uint256.Int).SetBytes(acc.Balance.Bytes()),
			Root:     root,
			CodeHash: keccak.Keccak(acc.Code).Bytes(),
		}

		if err := updateAccount(g.ParentStateTrie, addr, stateAcc); err != nil {
			return err
		}
	}
	expected := g.Block.Root()
	actual, err := g.ParentStateTrie.Hash()
	if err != nil {
		return err
	}
	if expected != actual {
		return fmt.Errorf(
			"block %d root mismatch: expected %#x, got %#x",
			g.Block.NumberU64(),
			expected,
			actual,
		)
	}
	return nil
}

func apply(
	vmConfig vm.Config,
	stateDB *state.StateDB,
	block *types.Block,
	txs types.Transactions,
	getHash func(uint64) common.Hash,
	chainConfig *params.ChainConfig,
) (*state.StateDB, error) {
	rnd := block.MixDigest()
	vmContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    block.Coinbase(),
		BlockNumber: block.Number(),
		Time:        block.Time(),
		Difficulty:  block.Difficulty(),
		GasLimit:    block.GasLimit(),
		GetHash:     getHash,
		BaseFee:     block.BaseFee(),
		Random:      &rnd,
	}
	var (
		gasPool    = new(core.GasPool)
		gasUsed    = uint64(0)
		txIndex    = 0
		invalidTxs types.Transactions
		validTxs   types.Transactions
		signer     = types.MakeSigner(chainConfig, block.Number(), block.Time())
	)
	gasPool.AddGas(block.GasLimit())
	rules := chainConfig.Rules(block.Number(), true, block.Time())

	for _, tx := range txs {
		isAnchor := txIndex == 0
		if isAnchor {
			if err := tx.MarkAsAnchor(); err != nil {
				return nil, err
			}
		}

		if tx.Type() == types.BlobTxType {
			if isAnchor {
				return nil, errors.New("anchor tx cannot be a blob tx")
			}
			log.Warn("Skip a blob transaction", "hash", tx.Hash())
			invalidTxs = append(invalidTxs, tx)
			continue
		}
		from, _ := types.Sender(signer, tx)
		stateDB.Prepare(
			rules,
			from,
			block.Coinbase(),
			tx.To(),
			vm.ActivePrecompiles(rules),
			tx.AccessList(),
		)
		stateDB.SetTxContext(tx.Hash(), txIndex)
		var (
			snapshot = stateDB.Snapshot()
			prevGas  = gasPool.Gas()
		)
		evm := vm.NewEVM(vmContext, stateDB, chainConfig, vmConfig)
		_, err := core.ApplyTransaction(evm, gasPool, stateDB, block.Header(), tx, &gasUsed)
		if err != nil {
			if isAnchor {
				return nil, err
			}
			log.Warn(
				"rejected tx",
				"index", txIndex,
				"hash", tx.Hash(),
				"from", from,
				"error", err,
			)
			stateDB.RevertToSnapshot(snapshot)
			gasPool.SetGas(prevGas)
			invalidTxs = append(invalidTxs, tx)
			continue
		}

		txIndex++
		validTxs = append(validTxs, tx)
	}
	stateDB.IntermediateRoot(chainConfig.IsEIP158(vmContext.BlockNumber))
	// Apply withdrawals
	for _, w := range block.Withdrawals() {
		// Amount is in gwei, turn into wei
		amount := new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(params.GWei))
		stateDB.AddBalance(
			w.Address,
			uint256.MustFromBig(amount),
			tracing.BalanceIncreaseWithdrawal,
		)
	}
	// Commit block
	root, err := stateDB.Commit(
		vmContext.BlockNumber.Uint64(),
		chainConfig.IsEIP158(vmContext.BlockNumber),
		chainConfig.IsCancun(vmContext.BlockNumber, vmContext.Time),
	)
	if err != nil {
		return nil, fmt.Errorf("could not commit state: %v", err)
	}
	_ = validTxs
	if len(invalidTxs) > 0 {
		log.Warn("invalid transactions", len(invalidTxs))
	}
	return state.New(root, stateDB.Database())
}
