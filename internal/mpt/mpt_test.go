package mpt

import (
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/gaiko/internal"
)

const intSize = 32 << (^uint(0) >> 63)

func TestTriePointerNoKeccak(t *testing.T) {
	cases := map[string]string{
		"do":    "verb",
		"dog":   "puppy",
		"doge":  "coin",
		"horse": "stallion",
	}
	for k, v := range cases {
		node := newMptNode(&leafNode{
			prefix: []byte(k),
			value:  []byte(v),
		})

		expected, err := rlp.EncodeToBytes(node)
		require.NoError(t, err)

		ref, err := node.ref()
		require.NoError(t, err)
		actual := ref.(bytesMptNodeRef)
		assert.Equal(t, expected, []byte(actual))
	}
}

func TestToEncodedPath(t *testing.T) {
	cases := [][]byte{
		{0x0a, 0x0b, 0x0c, 0x0d},
		{0x0a, 0x0b, 0x0c},
	}
	expected := [][]byte{
		{0x00, 0xab, 0xcd},
		{0x1a, 0xbc},
	}
	expectedLeaf := [][]byte{
		{0x20, 0xab, 0xcd},
		{0x3a, 0xbc},
	}
	for i, nibbles := range cases {
		assert.Equal(t, expected[i], toEncodedPath(nibbles, false))
		assert.Equal(t, expectedLeaf[i], toEncodedPath(nibbles, true))
	}
}

func TestLcp(t *testing.T) {
	cases := []struct {
		a   []byte
		b   []byte
		cpl int
	}{
		{[]byte{}, []byte{}, 0},
		{[]byte{0xa}, []byte{0xa}, 1},
		{[]byte{0xa, 0xb}, []byte{0xa, 0xc}, 1},
		{[]byte{0xa, 0xb}, []byte{0xa, 0xb}, 2},
		{[]byte{0xa, 0xb}, []byte{0xa, 0xb, 0xc}, 2},
		{[]byte{0xa, 0xb, 0xc}, []byte{0xa, 0xb, 0xc}, 3},
		{[]byte{0xa, 0xb, 0xc}, []byte{0xa, 0xb, 0xc, 0xd}, 3},
		{[]byte{0xa, 0xb, 0xc, 0xd}, []byte{0xa, 0xb, 0xc, 0xd}, 4},
	}
	for _, tc := range cases {
		assert.Equal(t, lcp(tc.a, tc.b), tc.cpl)
	}
}

func TestEmptyKey(t *testing.T) {
	trie := New()
	_, err := trie.Insert([]byte{}, []byte("empty"))
	require.NoError(t, err)
	data, err := trie.Get([]byte{})
	require.NoError(t, err)
	assert.Equal(t, []byte("empty"), data)
	ok, err := trie.Delete([]byte{})
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestClear(t *testing.T) {
	trie := New()
	_, err := trie.Insert([]byte("dog"), []byte("puppy"))
	require.NoError(t, err)
	assert.False(t, trie.IsEmpty())
	hash, err := trie.Hash()
	require.NoError(t, err)
	assert.NotEqual(t, hash, types.EmptyRootHash)
	trie.Clear()
	assert.True(t, trie.IsEmpty())
	hash, err = trie.Hash()
	require.NoError(t, err)
	assert.Equal(t, hash, types.EmptyRootHash)
}

func TestTiny(t *testing.T) {
	trie := New()
	_, err := trie.InsertRLP([]byte("a"), uint8(0))
	require.NoError(t, err)
	_, err = trie.InsertRLP([]byte("b"), uint8(1))
	require.NoError(t, err)
	assert.False(t, trie.IsEmpty())
	expected, err := hex.DecodeString("d816d680c3208180c220018080808080808080808080808080")
	require.NoError(t, err)
	ref, err := trie.ref()
	require.NoError(t, err)
	actual := ref.(bytesMptNodeRef)
	assert.Equal(t, expected, []byte(actual))
}

func TestPartial(t *testing.T) {
	trie := New()
	_, err := trie.InsertRLP([]byte("aa"), uint8(0))
	require.NoError(t, err)
	_, err = trie.InsertRLP([]byte("ab"), uint8(1))
	require.NoError(t, err)
	_, err = trie.InsertRLP([]byte("ba"), uint8(2))
	require.NoError(t, err)

	_, err = trie.Hash()
	require.NoError(t, err)

	// replace one node with its digest
	node, ok := trie.data.(*extensionNode)
	require.True(t, ok)
	newTrie := newMptNode(node)
	hash, err := newTrie.Hash()
	require.NoError(t, err)
	newTrie.data = (*digestNode)(&hash)
	require.True(t, newTrie.IsDigest())
}

func TestBranchValue(t *testing.T) {
	trie := New()
	_, err := trie.Insert([]byte("do"), []byte("verb"))
	require.NoError(t, err)
	_, err = trie.Insert([]byte("dog"), []byte("puppy"))
	require.Error(t, err)
}

func TestInsert(t *testing.T) {
	trie := New()
	vals := map[string]string{
		"painting": "place",
		"guest":    "ship",
		"mud":      "leave",
		"paper":    "call",
		"gate":     "boast",
		"tongue":   "gain",
		"baseball": "wait",
		"tale":     "lie",
		"mood":     "cope",
		"menu":     "fear",
	}
	for key, val := range vals {
		ok, err := trie.Insert([]byte(key), []byte(val))
		require.NoError(t, err)
		assert.True(t, ok)
	}
	expected := common.HexToHash("2bab6cdf91a23ebf3af683728ea02403a98346f99ed668eec572d55c70a4b08f")
	actual, err := trie.Hash()
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
	for key, val := range vals {
		data, err := trie.Get([]byte(key))
		require.NoError(t, err)
		assert.Equal(t, []byte(val), data)
	}
}

func keyFunc(i int) []byte {
	switch intSize {
	case 32:
		key := make([]byte, 4)
		binary.BigEndian.PutUint32(key, uint32(i))
		return key
	case 64:
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(i))
		return key
	}
	panic("unreachable")
}

func TestMpt(t *testing.T) {
	const N = 512
	trie := New()
	for i := 0; i < N; i++ {
		key := keyFunc(i)
		ok, err := trie.InsertRLP(internal.Keccak(key), uint(i))
		require.NoError(t, err)
		require.True(t, ok)

		ref := New()
		for j := i; j >= 0; j-- {
			key := keyFunc(j)
			ok, err := ref.InsertRLP(internal.Keccak(key), uint(j))
			require.NoError(t, err)
			require.True(t, ok)
		}
		expected, err := trie.Hash()
		require.NoError(t, err)
		actual, err := ref.Hash()
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	actual := common.HexToHash("7310027edebdd1f7c950a7fb3413d551e85dff150d45aca4198c2f6315f9b4a7")
	expected, err := trie.Hash()
	require.NoError(t, err)
	require.Equal(t, expected, actual)
	for i := 0; i < N; i++ {
		key := keyFunc(i)
		data, err := trie.Get(internal.Keccak(key))
		require.NoError(t, err)
		var val uint64
		err = rlp.DecodeBytes(data, &val)
		require.NoError(t, err)
		require.Equal(t, uint64(i), val)
	}
}

func TestKeccak(t *testing.T) {
	key := keyFunc(1)
	expected := internal.Keccak(key)
	actual := []byte{
		0x6c, 0x31, 0xfc, 0x15, 0x42, 0x2e, 0xba, 0xd2,
		0x8a, 0xaf, 0x90, 0x89, 0xc3, 0x6, 0x70, 0x2f,
		0x67, 0x54, 0xb, 0x53, 0xc7, 0xee, 0xa8, 0xb7,
		0xd2, 0x94, 0x10, 0x44, 0xb0, 0x27, 0x10, 0xf,
	}
	assert.Equal(t, expected, actual)
}

func TestIndexTrie(t *testing.T) {
	const N = 512

	trie := New()
	// insert
	for i := 0; i < N; i++ {
		key, err := rlp.EncodeToBytes(uint(i))
		require.NoError(t, err)
		ok, err := trie.InsertRLP(key, uint(i))
		require.NoError(t, err)
		require.True(t, ok)

		ref := New()
		for j := i; j >= 0; j-- {
			key, err := rlp.EncodeToBytes(uint(j))
			require.NoError(t, err)

			ok, err := ref.InsertRLP(key, uint(j))
			require.NoError(t, err)
			require.True(t, ok)
		}
		expected, err := trie.Hash()
		require.NoError(t, err)
		actual, err := ref.Hash()
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	// get
	for i := 0; i < N; i++ {
		key, err := rlp.EncodeToBytes(uint(i))
		require.NoError(t, err)
		data, err := trie.Get(key)
		require.NoError(t, err)
		var val uint
		err = rlp.DecodeBytes(data, &val)
		require.NoError(t, err)
		require.Equal(t, uint(i), val)

		key, err = rlp.EncodeToBytes(uint(i + N))
		require.NoError(t, err)
		notFound, err := trie.Get(key)
		require.NoError(t, err)
		require.Nil(t, notFound)
	}
	// delete
	for i := 0; i < N; i++ {
		key, err := rlp.EncodeToBytes(uint(i))
		require.NoError(t, err)
		ok, err := trie.Delete(key)
		require.NoError(t, err)
		require.True(t, ok)

		ref := New()
		for j := N - 1; j >= i+1; j-- {
			key, err := rlp.EncodeToBytes(uint(j))
			require.NoError(t, err)
			ok, err := ref.InsertRLP(key, uint(j))
			require.NoError(t, err)
			require.True(t, ok)
		}
		expected, err := trie.Hash()
		require.NoError(t, err)
		actual, err := ref.Hash()
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	assert.True(t, trie.IsEmpty())
}
