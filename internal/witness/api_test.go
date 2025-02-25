package witness

import (
	"fmt"
	"testing"
)

func TestABI(t *testing.T) {
	for _, input := range batchProposedEvent.Inputs {
		fmt.Printf("%#v\n", input)
		fmt.Println()
	}
}
