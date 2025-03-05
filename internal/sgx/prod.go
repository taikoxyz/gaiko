//go:build !test

package sgx

import "github.com/taikoxyz/gaiko/internal/flags"

func NewProvider(args *flags.Arguments) Provider {
	switch args.SGXType {
	case flags.GramineSGXType:
		return NewGramineProvider(args)
	default:
		return NewEgoProvider(args)
	}
}
