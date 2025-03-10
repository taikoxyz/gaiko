package prover

import (
	"context"
)

type Prover interface {
	Oneshot(ctx context.Context) error
	BatchOneshot(ctx context.Context) error
	Aggregate(ctx context.Context) error
	Bootstrap(ctx context.Context) error
	Check(ctx context.Context) error
}
