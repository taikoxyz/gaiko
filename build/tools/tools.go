//go:build tools

package tools

import (
	// Ignore any error here as the import is only needed for go:generate.
	_ "github.com/fjl/gencodec"
)
