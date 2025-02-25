package sgx

import (
	"context"

	"github.com/taikoxyz/gaiko/internal/flags"
)

type Prover interface {
	Oneshot(ctx context.Context, args *flags.Arguments) (*Proof, error)
	Aggregate(ctx context.Context, args *flags.Arguments) (*Proof, error)
	Bootstrap(ctx context.Context, args *flags.Arguments) error
	Check(ctx context.Context) error
}

type Proof struct {
	Proof           string `json:"proof"`
	Quote           string `json:"quote"`
	PublicKey       string `json:"public_key"`
	InstanceAddress string `json:"instance_address"`
	Input           string `json:"input"`
}
