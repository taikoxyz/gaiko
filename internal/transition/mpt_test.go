package transition

import (
	"encoding/binary"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestMpt(t *testing.T) {
	keyFunc := func(i int) []byte {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(i))
		return key
	}
	trie := NewEmptyMptNode()
	for i := 0; i < 512; i++ {
		key := keyFunc(i)
		ok, err := trie.InsertRLP(keccak(key), uint64(i))
		assert.True(t, ok)
		assert.NoError(t, err)

		ref := NewEmptyMptNode()
		for j := i; j >= 0; j-- {
			key := keyFunc(j)
			ok, err := ref.InsertRLP(keccak(key), uint64(j))
			assert.True(t, ok)
			assert.NoError(t, err)
		}
		expected, err := trie.Hash()
		assert.NoError(t, err)
		got, err := ref.Hash()
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	}
	actual := common.HexToHash("7310027edebdd1f7c950a7fb3413d551e85dff150d45aca4198c2f6315f9b4a7")
	expected, err := trie.Hash()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
	for i := 0; i < 512; i++ {
		key := keyFunc(i)
		data, err := trie.Get(key)
		assert.NoError(t, err)
		var val uint64
		err = rlp.DecodeBytes(data, &val)
		assert.NoError(t, err)
		assert.Equal(t, uint64(i), val)
	}
}
