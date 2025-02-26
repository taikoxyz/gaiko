package prover

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/keccak"
	"github.com/taikoxyz/gaiko/internal/sgx"
	"github.com/taikoxyz/gaiko/internal/witness"
)

var addr32Padding [32 - common.AddressLength]byte

type SGXProver struct {
	provider sgx.Provider
	args     *flags.Arguments
}

var _ Prover = (*SGXProver)(nil)

func NewSGXProver(args *flags.Arguments) *SGXProver {
	return &SGXProver{
		args:     args,
		provider: sgx.NewProvider(args),
	}
}

func (p *SGXProver) Oneshot(
	ctx context.Context,
) (*ProofResponse, error) {
	var driver witness.GuestInput
	return genSGXProof(ctx, p.args, &driver, p.provider)
}

func (p *SGXProver) BatchOneshot(
	ctx context.Context,
) (*ProofResponse, error) {
	var driver witness.BatchGuestInput
	return genSGXProof(ctx, p.args, &driver, p.provider)
}

func (p *SGXProver) Aggregate(
	ctx context.Context,
) (*ProofResponse, error) {
	prevPrivKey, err := p.provider.LoadPrivateKey()
	if err != nil {
		return nil, err
	}
	newInstance := crypto.PubkeyToAddress(prevPrivKey.PublicKey)
	var input witness.RawAggregationGuestInput
	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		return nil, err
	}
	oldInstance := common.BytesToAddress(input.Proofs[0].Proof[4:24])
	curInstance := oldInstance
	for i, proof := range input.Proofs {
		pubKey, err := crypto.SigToPub(proof.Input.Bytes(), proof.Proof[24:])
		if err != nil {
			return nil, err
		}
		if crypto.PubkeyToAddress(*pubKey) != curInstance {
			return nil, fmt.Errorf("invalid proof[%d]", i)
		}
		curInstance = common.BytesToAddress(proof.Proof[4:24])
	}
	if newInstance != curInstance {
		return nil, fmt.Errorf("invalid instance: %s", curInstance)
	}

	aggOutputCombine := make([]byte, 0, (len(input.Proofs)+2)*32)
	aggOutputCombine = append(aggOutputCombine, addr32Padding[:]...)
	aggOutputCombine = append(aggOutputCombine, oldInstance.Bytes()...)
	aggOutputCombine = append(aggOutputCombine, addr32Padding[:]...)
	aggOutputCombine = append(aggOutputCombine, newInstance.Bytes()...)
	for _, proof := range input.Proofs {
		aggOutputCombine = append(aggOutputCombine, proof.Input.Bytes()...)
	}

	aggHash := keccak.Keccak(aggOutputCombine)
	sig, err := crypto.Sign(aggHash, prevPrivKey)
	if err != nil {
		return nil, err
	}

	proof := NewAggregateProof(p.args.InstanceID, oldInstance, newInstance, sig)
	quote, err := p.provider.LoadQuote(newInstance)
	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		Proof:           proof[:],
		Quote:           quote,
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           common.BytesToHash(aggHash),
	}, nil
}

func (p *SGXProver) Bootstrap(ctx context.Context) error {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	err = p.provider.SavePrivateKey(privKey)
	if err != nil {
		return err
	}
	fmt.Printf("Public key: %x\n", privKey.PublicKey)
	newInstance := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("Instance address: %x\n", newInstance)

	quote, err := p.provider.LoadQuote(newInstance)
	if err != nil {
		return err
	}
	b := &sgx.BootstrapData{
		PublicKey:   crypto.FromECDSAPub(&privKey.PublicKey),
		NewInstance: newInstance,
		Quote:       quote,
	}

	return p.provider.SaveBootstrap(b)
}

func (p *SGXProver) Check(ctx context.Context) error {
	_, err := p.provider.LoadPrivateKey()
	return err
}
