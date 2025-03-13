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
}

var _ Prover = (*SGXProver)(nil)

func NewSGXProver(args *flags.Arguments) *SGXProver {
	return &SGXProver{
		sgxProvider: tee.NewSGXProvider(args),
	}
}

func (p *SGXProver) Oneshot(ctx context.Context, args *flags.Arguments) error {
	var input witness.GuestInput
	return genOneshotProof(ctx, args, &input, p.sgxProvider)
}

func (p *SGXProver) BatchOneshot(ctx context.Context, args *flags.Arguments) error {
	var input witness.BatchGuestInput
	return genOneshotProof(ctx, args, &input, p.sgxProvider)
}

func (p *SGXProver) Aggregate(ctx context.Context, args *flags.Arguments) error {
	return genAggregateProof(ctx, args, p.sgxProvider)
}

func (p *SGXProver) Bootstrap(ctx context.Context, args *flags.Arguments) error {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	err = p.sgxProvider.SavePrivateKey(args, privKey)
	if err != nil {
		return err
	}
	fmt.Printf("Public key: %#x\n", privKey.PublicKey)
	newInstance := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("Instance address: %#x\n", newInstance)

	quote, err := p.sgxProvider.LoadQuote(args, newInstance)
	if err != nil {
		return err
	}
	b := &tee.BootstrapData{
		PublicKey:   crypto.FromECDSAPub(&privKey.PublicKey),
		NewInstance: newInstance,
		Quote:       quote.Bytes(),
	}

	return p.sgxProvider.SaveBootstrap(args, b)
}

func (p *SGXProver) Check(ctx context.Context, args *flags.Arguments) error {
	_, err := p.sgxProvider.LoadPrivateKey(args)
	return err
}
