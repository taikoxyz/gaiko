package transition

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/taikoxyz/gaiko/internal"
	"github.com/urfave/cli/v2"
)

func Oneshot(ctx *cli.Context) error {
	var driver GuestDriver
	err := json.NewDecoder(os.Stdin).Decode(&driver)
	if err != nil {
		return err
	}
	chainConfig, err := driver.ChainConfig()
	if err != nil {
		return err
	}
	for pair := range driver.GuestInputs() {
		g := pair.GuestInput
		txs := pair.Transactions
		preState, err := g.makePreState()
		if err != nil {
			return err
		}
		statedb, err := apply(vm.Config{}, preState.statedb, g.Block, txs, preState.getHash, chainConfig)
		if err != nil {
			return err
		}
		collector := make(Dumper)
		statedb.DumpToCollector(collector, nil)
		for addr, _ := range preState.accounts {
			_, ok := collector[addr]
			if !ok {
				// Account is deleted
				key := internal.Keccak(addr.Bytes())
				if _, err := g.ParentStateTrie.Delete(key); err != nil {
					return err
				}
			}
		}

		for addr, acc := range collector {
			storageEntry, ok := g.ParentStorage[addr]
			if !ok {
				return fmt.Errorf("account not found for address: %s", addr.Hex())
			}
			_, ok = preState.accounts[addr]
			if !ok {
				// New Account
				storageEntry.Trie.Clear()
			}
			for slot, value := range acc.Storage {
				key := internal.Keccak(slot.Bytes())
				if value == (common.Hash{}) {
					if _, err := storageEntry.Trie.Delete(key); err != nil {
						return err
					}
				} else {
					if err := updateStorage(storageEntry.Trie, slot.Bytes(), value.Bytes()); err != nil {
						return err
					}
				}
			}
			root, err := storageEntry.Trie.Hash()
			if err != nil {
				return err
			}
			stateAcc := &types.StateAccount{
				Nonce:    acc.Nonce,
				Balance:  new(uint256.Int).SetBytes(acc.Balance.Bytes()),
				Root:     root,
				CodeHash: internal.Keccak(acc.Code),
			}

			if err := updateAccount(g.ParentStateTrie, addr, stateAcc); err != nil {
				return err
			}
		}
	}
	return nil
}

func apply(vmConfig vm.Config, statedb *state.StateDB, block *types.Block, txs types.Transactions, getHash func(uint64) common.Hash, chainConfig *params.ChainConfig) (*state.StateDB, error) {
	var (
		signer  = types.MakeSigner(chainConfig, new(big.Int).SetUint64(block.NumberU64()), block.Time())
		gaspool = new(core.GasPool)
		gasUsed = uint64(0)
		txIndex = 0
	)

	gaspool.AddGas(block.GasLimit())

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
			continue
		}
		msg, err := core.TransactionToMessage(tx, signer, block.BaseFee())
		if err != nil {
			if isAnchor {
				return nil, err
			}
			log.Warn("rejected tx", "index", txIndex, "hash", tx.Hash(), "error", err)
			continue
		}
		statedb.SetTxContext(tx.Hash(), txIndex)
		var (
			txContext = core.NewEVMTxContext(msg)
			snapshot  = statedb.Snapshot()
			prevGas   = gaspool.Gas()
		)
		evm := vm.NewEVM(vmContext, txContext, statedb, chainConfig, vmConfig)

		// (ret []byte, usedGas uint64, failed bool, err error)
		msgResult, err := core.ApplyMessage(evm, msg, gaspool)
		if err != nil {
			if isAnchor {
				return nil, err
			}
			log.Warn("rejected tx", "index", txIndex, "hash", tx.Hash(), "from", msg.From, "error", err)
			statedb.RevertToSnapshot(snapshot)
			gaspool.SetGas(prevGas)
			continue
		}

		gasUsed += msgResult.UsedGas
		if chainConfig.IsByzantium(vmContext.BlockNumber) {
			statedb.Finalise(true)
		} else {
			statedb.IntermediateRoot(chainConfig.IsEIP158(vmContext.BlockNumber))
		}
		txIndex++
	}
	statedb.IntermediateRoot(chainConfig.IsEIP158(vmContext.BlockNumber))

	// Apply withdrawals
	for _, w := range block.Withdrawals() {
		// Amount is in gwei, turn into wei
		amount := new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(params.GWei))
		statedb.AddBalance(w.Address, uint256.MustFromBig(amount), tracing.BalanceIncreaseWithdrawal)
	}
	// Commit block
	root, err := statedb.Commit(vmContext.BlockNumber.Uint64(), chainConfig.IsEIP158(vmContext.BlockNumber))
	if err != nil {
		return nil, fmt.Errorf("could not commit state: %v", err)
	}

	return state.New(root, statedb.Database())
}
