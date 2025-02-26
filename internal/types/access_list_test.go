package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessListUnmarshal(t *testing.T) {
	inputs := []struct {
		name string
		data string
	}{
		{
			name: "empty",
			data: `[]`,
		},
		{
			name: "single",
			data: `[{"address":"0x3262f13a39efaca789ae58390441c9ed76bc658a", "storage_keys": ["0x47a89ac076877652cf199a0653214e97594c5138a173167fceb6017bcee2c588"]}]`,
		},

		{
			name: "null storage keys",
			data: `[{"address":"0x3262f13a39efaca789ae58390441c9ed76bc658a", "storage_keys": null}]`,
		},
	}

	for idx, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			var accessList AccessList
			err := json.Unmarshal([]byte(input.data), &accessList)
			require.NoError(t, err)
			switch idx {
			case 0:
				assert.Equal(t, len(accessList), 0)
			case 1:
				assert.Equal(t, len(accessList), 1)
				assert.Equal(t, len(accessList[0].StorageKeys), 1)
			case 2:
				assert.Equal(t, len(accessList), 1)
				assert.Equal(t, len(accessList[0].StorageKeys), 0)
			}
		})
	}
}
