package types

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

func TestEmptyHash(t *testing.T) {
	fmt.Printf("%#v\n", types.EmptyReceiptsHash.Bytes())
}
