//go:build !dev

package tee

import "github.com/taikoxyz/gaiko/internal/flags"

func NewSGXProvider(args *flags.Arguments) Provider {
	switch args.SGXType {
	case flags.GramineSGXType:
		return NewGramineProvider(args)
	default:
		return NewEgoProvider(args)
	}
}
