package mpt

import (
	"encoding/binary"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/internal"
)

const intSize = 32 << (^uint(0) >> 63)

func TestMpt(t *testing.T) {
	n := 512
	keyFunc := func(i int) []byte {
		switch intSize {
		case 32:
			key := make([]byte, 32)
			binary.BigEndian.PutUint32(key, uint32(i))
			return key
		case 64:
			key := make([]byte, 64)
			binary.BigEndian.PutUint64(key, uint64(i))
			return key
		}
		return nil
	}
	trie := NewEmptyMptNode()
	for i := 0; i < n; i++ {
		key := keyFunc(i)
		ok, err := trie.InsertRLP(internal.Keccak(key), uint(i))
		require.NoError(t, err)
		assert.True(t, ok)

		ref := NewEmptyMptNode()
		for j := i; j >= 0; j-- {
			key := keyFunc(j)
			ok, err := ref.InsertRLP(internal.Keccak(key), uint(j))
			require.NoError(t, err)
			assert.True(t, ok)
		}
		expected, err := trie.Hash()
		require.NoError(t, err)
		actual, err := ref.Hash()
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
	actual := common.HexToHash("7310027edebdd1f7c950a7fb3413d551e85dff150d45aca4198c2f6315f9b4a7")
	expected, err := trie.Hash()
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
	for i := 0; i < n; i++ {
		key := keyFunc(i)
		data, err := trie.Get(internal.Keccak(key))
		require.NoError(t, err)
		var val uint64
		err = rlp.DecodeBytes(data, &val)
		require.NoError(t, err)
		assert.Equal(t, uint64(i), val)
	}
}
