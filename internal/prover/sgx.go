package prover

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/tee"
	"github.com/taikoxyz/gaiko/internal/witness"
)

var addr2HashPadding [common.HashLength - common.AddressLength]byte

type SGXProver struct {
	sgxProvider tee.Provider
	args        *flags.Arguments
}

var _ Prover = (*SGXProver)(nil)

func NewSGXProver(args *flags.Arguments) *SGXProver {
	return &SGXProver{
		args:        args,
		sgxProvider: tee.NewSGXProvider(args),
	}
}

func (p *SGXProver) Oneshot(
	ctx context.Context,
) error {
	var driver witness.GuestInput
	proof, err := genOneshotProof(ctx, p.args, &driver, p.sgxProvider)
	if err != nil {
		return err
	}
	return proof.Output(p.args.ProofWriter)
}

func (p *SGXProver) BatchOneshot(
	ctx context.Context,
) error {
	var driver witness.BatchGuestInput
	proof, err := genOneshotProof(ctx, p.args, &driver, p.sgxProvider)
	if err != nil {
		return err
	}
	return proof.Output(p.args.ProofWriter)
}

func (p *SGXProver) Aggregate(
	ctx context.Context,
) error {
	proof, err := genAggregateProof(ctx, p.args, p.sgxProvider)
	if err != nil {
		return err
	}
	return proof.Output(p.args.ProofWriter)
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
	fmt.Printf("Public key: %#x\n", privKey.PublicKey)
	newInstance := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("Instance address: %#x\n", newInstance)

	quote, err := p.sgxProvider.LoadQuote(newInstance)
	if err != nil {
		return err
	}
	b := &tee.BootstrapData{
		PublicKey:   crypto.FromECDSAPub(&privKey.PublicKey),
		NewInstance: newInstance,
		Quote:       quote.Bytes(),
	}

	return p.sgxProvider.SaveBootstrap(b)
}

func (p *SGXProver) Check(ctx context.Context) error {
	_, err := p.sgxProvider.LoadPrivateKey()
	return err
}
