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
	sgxProvider sgx.Provider
	args        *flags.Arguments
}

var _ Prover = (*SGXProver)(nil)

func NewSGXProver(args *flags.Arguments) *SGXProver {
	return &SGXProver{
		args:        args,
		sgxProvider: sgx.NewProvider(args),
	}
}

func (p *SGXProver) Oneshot(
	ctx context.Context,
) (*ProofResponse, error) {
	var driver witness.GuestInput
	return genOneshotProof(ctx, p.args, &driver, p.sgxProvider)
}

func (p *SGXProver) BatchOneshot(
	ctx context.Context,
) (*ProofResponse, error) {
	var driver witness.BatchGuestInput
	return genOneshotProof(ctx, p.args, &driver, p.sgxProvider)
}

func (p *SGXProver) Aggregate(
	ctx context.Context,
) (*ProofResponse, error) {
	prevPrivKey, err := p.sgxProvider.LoadPrivateKey()
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

	combinedHashes := make([]byte, 0, (len(input.Proofs)+2)*32)
	combinedHashes = append(combinedHashes, addr32Padding[:]...)
	combinedHashes = append(combinedHashes, oldInstance.Bytes()...)
	combinedHashes = append(combinedHashes, addr32Padding[:]...)
	combinedHashes = append(combinedHashes, newInstance.Bytes()...)
	for _, proof := range input.Proofs {
		combinedHashes = append(combinedHashes, proof.Input.Bytes()...)
	}

	aggHash := keccak.Keccak(combinedHashes)
	sign, err := crypto.Sign(aggHash.Bytes(), prevPrivKey)
	if err != nil {
		return nil, err
	}

	proof := NewAggregateProof(p.args.SGXInstanceID, oldInstance, newInstance, sign)
	quote, err := p.sgxProvider.LoadQuote(newInstance)
	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		Proof:           proof[:],
		Quote:           quote,
		PublicKey:       crypto.FromECDSAPub(&prevPrivKey.PublicKey),
		InstanceAddress: newInstance,
		Input:           aggHash,
	}, nil
}

func (p *SGXProver) Bootstrap(ctx context.Context) error {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	err = p.sgxProvider.SavePrivateKey(privKey)
	if err != nil {
		return err
	}
	fmt.Printf("Public key: %x\n", privKey.PublicKey)
	newInstance := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("Instance address: %x\n", newInstance)

	quote, err := p.sgxProvider.LoadQuote(newInstance)
	if err != nil {
		return err
	}
	b := &sgx.BootstrapData{
		PublicKey:   crypto.FromECDSAPub(&privKey.PublicKey),
		NewInstance: newInstance,
		Quote:       quote,
	}

	return p.sgxProvider.SaveBootstrap(b)
}

func (p *SGXProver) Check(ctx context.Context) error {
	_, err := p.sgxProvider.LoadPrivateKey()
	return err
}
