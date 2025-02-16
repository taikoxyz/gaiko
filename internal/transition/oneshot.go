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
	"github.com/urfave/cli/v2"
)

func Oneshot(ctx *cli.Context) error {
	var driver Driver
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
		statedb, err := g.apply(vm.Config{}, preState.statedb, txs, preState.getHash, chainConfig)
		if err != nil {
			return err
		}
		collector := make(Dumper)
		statedb.DumpToCollector(collector, nil)
		for addr, _ := range preState.accounts {
			_, ok := collector[addr]
			if !ok {
				// Account is deleted
				key := keccak(addr.Bytes())
				if err := g.ParentStateTrie.Delete(key); err != nil {
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
				storageEntry.Trie.Reset()
			}
			for slot, value := range acc.Storage {
				key := keccak(slot.Bytes())
				if value == (common.Hash{}) {
					if err := storageEntry.Trie.Delete(key); err != nil {
						return err
					}
				} else {
					if err := updateStorage(&storageEntry.Trie, slot.Bytes(), value.Bytes()); err != nil {
						return err
					}
				}
			}
			stateAcc := &types.StateAccount{
				Nonce:    acc.Nonce,
				Balance:  new(uint256.Int).SetBytes(acc.Balance.Bytes()),
				Root:     storageEntry.Trie.Hash(),
				CodeHash: keccak(acc.Code),
			}

			if err := updateAccount(&g.ParentStateTrie, addr, stateAcc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *GuestInput) apply(vmConfig vm.Config, statedb *state.StateDB, txs types.Transactions, getHash func(uint64) common.Hash, chainConfig *params.ChainConfig) (*state.StateDB, error) {
	var (
		signer  = types.MakeSigner(chainConfig, new(big.Int).SetUint64(g.Block.NumberU64()), g.Block.Time())
		gaspool = new(core.GasPool)
		gasUsed = uint64(0)
		txIndex = 0
	)

	gaspool.AddGas(g.Block.GasLimit())

	rnd := g.Block.MixDigest()
	vmContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    g.Block.Coinbase(),
		BlockNumber: g.Block.Number(),
		Time:        g.Block.Time(),
		Difficulty:  g.Block.Difficulty(),
		GasLimit:    g.Block.GasLimit(),
		GetHash:     getHash,
		BaseFee:     g.Block.BaseFee(),
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
		msg, err := core.TransactionToMessage(tx, signer, g.Block.BaseFee())
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
	for _, w := range g.Block.Withdrawals() {
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
