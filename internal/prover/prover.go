package prover

import (
	"context"
)

type Prover interface {
	Oneshot(ctx context.Context) (*ProofResponse, error)
	BatchOneshot(ctx context.Context) (*ProofResponse, error)
	Aggregate(ctx context.Context) (*ProofResponse, error)
	Bootstrap(ctx context.Context) error
	Check(ctx context.Context) error
}
