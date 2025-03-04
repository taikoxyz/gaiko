package mpt

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/taikoxyz/gaiko/internal/keccak"
)

type MptNode struct {
	data      mptNodeData
	cachedRef mptNodeRef
}

func newMptNode(data mptNodeData) *MptNode {
	return &MptNode{
		data: data,
	}
}

// New creates a new empty MPT node.
func New() *MptNode {
	return newMptNode(&nullNode{})
}

// Clear resets the node to an empty state.
func (m *MptNode) Clear() {
	m.data = &nullNode{}
	m.cachedRef = nil
}

// Hash returns the Keccak-256 hash of the node.
// For null nodes, it returns the EmptyRootHash.
// For other nodes, it computes the hash from the node reference.
func (m *MptNode) Hash() (common.Hash, error) {
	_, ok := m.data.(*nullNode)
	if ok {
		return types.EmptyRootHash, nil
	}
	ref, err := m.ref()
	if err != nil {
		return common.Hash{}, err
	}
	return ref.hash(), nil
}

// IsEmpty returns true if the node is a null node.
func (m *MptNode) IsEmpty() bool {
	_, ok := m.data.(*nullNode)
	return ok
}

// IsDigest returns true if the node is a digest node.
func (m *MptNode) IsDigest() bool {
	_, ok := m.data.(*digestNode)
	return ok
}

// Nibs returns the nibble-encoded path prefix of the node.
// Returns nil for null, branch, and digest nodes.
func (m *MptNode) Nibs() []byte {
	switch data := m.data.(type) {
	case *nullNode, *branchNode, *digestNode:
		return nil
	case *leafNode:
		return prefixNibs(data.prefix)
	case *extensionNode:
		return prefixNibs(data.prefix)
	default:
		return nil
	}
}

// Get retrieves a value from the trie by its key.
// Returns nil if the key does not exist in the trie.
func (m *MptNode) Get(key []byte) ([]byte, error) {
	return m.get(toNibs(key))
}

// Delete removes a key-value pair from the trie.
// Returns true if the key was successfully deleted, false if the key wasn't found.
func (m *MptNode) Delete(key []byte) (bool, error) {
	return m.delete(toNibs(key))
}

// Insert adds or updates a key-value pair in the trie.
// Returns true if the key was added or modified, false otherwise.
func (m *MptNode) Insert(key []byte, value []byte) (bool, error) {
	return m.insert(toNibs(key), value)
}

// InsertRLP encodes the provided value using RLP encoding and inserts it into the trie.
// Returns true if the key was added or modified, false otherwise.
func (m *MptNode) InsertRLP(key []byte, value any) (bool, error) {
	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		return false, err
	}
	return m.insert(toNibs(key), data)
}

func (m *MptNode) insert(keyNibs []byte, value []byte) (bool, error) {
	if len(value) == 0 {
		panic("value must not be empty")
	}
	switch data := m.data.(type) {
	case *nullNode:
		m.data = &leafNode{
			prefix: toEncodedPath(keyNibs, true),
			value:  value,
		}
	case *branchNode:
		if len(keyNibs) == 0 {
			return false, errors.New("branch node with value")
		}
		idx, tail := keyNibs[0], keyNibs[1:]
		child := data[idx]
		if child != nil {
			ok, err := child.insert(tail, value)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		} else {
			data[idx] = newMptNode(&leafNode{
				prefix: toEncodedPath(tail, true),
				value:  value,
			})
		}
	case *leafNode:
		selfNibs := prefixNibs(data.prefix)
		commonLen := lcp(selfNibs, keyNibs)
		if commonLen == len(selfNibs) && commonLen == len(keyNibs) {
			if slices.Equal(data.value, value) {
				return false, nil
			}
			data.value = value
		} else if commonLen == len(selfNibs) || commonLen == len(keyNibs) {
			return false, errors.New("branch node with value")
		} else {
			splitPoint := commonLen + 1
			branch := &branchNode{}
			branch[selfNibs[commonLen]] = newMptNode(&leafNode{
				prefix: toEncodedPath(selfNibs[splitPoint:], true),
				value:  data.value,
			})
			branch[keyNibs[commonLen]] = newMptNode(&leafNode{
				prefix: toEncodedPath(keyNibs[splitPoint:], true),
				value:  value,
			})
			if commonLen > 0 {
				m.data = &extensionNode{
					prefix: toEncodedPath(selfNibs[:commonLen], false),
					child:  newMptNode(branch),
				}
			} else {
				m.data = branch
			}
		}
	case *extensionNode:
		selfNibs := prefixNibs(data.prefix)
		commonLen := lcp(selfNibs, keyNibs)
		if commonLen == len(selfNibs) {
			ok, err := data.child.insert(keyNibs[commonLen:], value)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		} else if commonLen == len(keyNibs) {
			return false, errors.New("branch node with value")
		} else {
			splitPoint := commonLen + 1
			branch := &branchNode{}
			if splitPoint < len(selfNibs) {
				branch[selfNibs[commonLen]] = newMptNode(&extensionNode{
					prefix: toEncodedPath(selfNibs[splitPoint:], false),
					child:  data.child,
				})
			} else {
				branch[selfNibs[commonLen]] = data.child
			}
			branch[keyNibs[commonLen]] = newMptNode(&leafNode{
				prefix: toEncodedPath(keyNibs[splitPoint:], true),
				value:  value,
			})
			if commonLen > 0 {
				m.data = &extensionNode{
					prefix: toEncodedPath(selfNibs[:commonLen], false),
					child:  newMptNode(branch),
				}
			} else {
				m.data = branch
			}
		}
	case *digestNode:
		return false, fmt.Errorf("node not resolved: %s", data)
	}
	m.cachedRef = nil
	return true, nil
}

func lcp(a, b []byte) int {
	minLen := min(len(a), len(b))

	for i := range minLen {
		if a[i] != b[i] {
			return i
		}
	}

	return minLen
}

func (m *MptNode) delete(keyNibs []byte) (bool, error) {
	switch data := m.data.(type) {
	case *nullNode:
		return false, nil
	case *branchNode:
		if len(keyNibs) == 0 {
			return false, errors.New("branch node with value")
		}
		idx, tail := keyNibs[0], keyNibs[1:]
		child := data[idx]
		if child == nil {
			return false, nil
		}
		ok, err := child.delete(tail)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		if child.IsEmpty() {
			data[idx] = nil
		}
		var (
			nextIdx   int
			nextChild *MptNode
			remaining int
		)
		for idx, child := range data {
			if child != nil {
				remaining += 1
				nextIdx = idx
				nextChild = child
			}
		}
		if remaining == 0 {
			panic("at least one remaining node")
		}
		// if there is only exactly one node left, we need to convert the branch
		if remaining == 1 {
			switch data := nextChild.data.(type) {
			case *leafNode:
				newNibs := slices.Concat([]byte{uint8(nextIdx)}, prefixNibs(data.prefix))
				m.data = &leafNode{
					prefix: toEncodedPath(newNibs, true),
					value:  data.value,
				}
			case *extensionNode:
				newNibs := slices.Concat([]byte{uint8(nextIdx)}, prefixNibs(data.prefix))

				m.data = &extensionNode{
					prefix: toEncodedPath(newNibs, false),
					child:  data.child,
				}
			case *branchNode, *digestNode:
				m.data = &extensionNode{
					prefix: toEncodedPath([]byte{byte(nextIdx)}, false),
					child:  nextChild,
				}
			case *nullNode:
				panic("unreachable")
			}
		}
	case *leafNode:
		if !slices.Equal(prefixNibs(data.prefix), keyNibs) {
			return false, nil
		}
		m.data = &nullNode{}
	case *extensionNode:
		selfNibs := prefixNibs(data.prefix)
		tail := stripPrefix(keyNibs, selfNibs)
		if tail != nil {
			ok, err := data.child.delete(tail)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		} else {
			return false, nil
		}

		switch data := data.child.data.(type) {
		case *nullNode:
			m.data = &nullNode{}
		case *leafNode:
			selfNibs = slices.Concat(selfNibs, prefixNibs(data.prefix))
			m.data = &leafNode{
				prefix: toEncodedPath(selfNibs, true),
				value:  data.value,
			}
		case *extensionNode:
			selfNibs = slices.Concat(selfNibs, prefixNibs(data.prefix))

			m.data = &extensionNode{
				prefix: toEncodedPath(selfNibs, false),
				child:  data.child,
			}
		case *branchNode, *digestNode:
		}
	case *digestNode:
		return false, fmt.Errorf("node not resolved: %s", data)
	}
	m.cachedRef = nil
	return true, nil
}

func (m *MptNode) get(keyNibs []byte) ([]byte, error) {
	switch data := m.data.(type) {
	case *nullNode:
		return nil, nil
	case *branchNode:
		if len(keyNibs) == 0 {
			return nil, nil
		}
		idx, tail := keyNibs[0], keyNibs[1:]
		if data[idx] == nil {
			return nil, nil
		}
		return data[idx].get(tail)
	case *leafNode:
		if bytes.Equal(prefixNibs(data.prefix), keyNibs) {
			return data.value, nil
		}
		return nil, nil
	case *extensionNode:
		prefix := prefixNibs(data.prefix)
		return data.child.get(stripPrefix(keyNibs, prefix))
	case *digestNode:
		return nil, fmt.Errorf("node not resolved: %s", data)
	}
	return nil, nil
}

func (m *MptNode) refEncode(w rlp.EncoderBuffer) error {
	ref, err := m.ref()
	if err != nil {
		return err
	}
	return ref.encodeRLP(w)
}

func (m *MptNode) ref() (mptNodeRef, error) {
	if m.cachedRef == nil {
		switch data := m.data.(type) {
		case *nullNode:
			m.cachedRef = bytesMptNodeRef(rlp.EmptyString)
		case *digestNode:
			m.cachedRef = digestMptNodeRef(*data)
		default:
			encoded, err := rlp.EncodeToBytes(m)
			if err != nil {
				return nil, err
			}
			if len(encoded) < 32 {
				m.cachedRef = bytesMptNodeRef(encoded)
			} else {
				m.cachedRef = digestMptNodeRef(keccak.Keccak(encoded))
			}
		}
	}
	return m.cachedRef, nil
}

func stripPrefix(nibs []byte, prefix []byte) []byte {
	if bytes.HasPrefix(nibs, prefix) {
		return nibs[len(prefix):]
	}
	return nil
}

func boolToInt(b bool) (n int) {
	if b {
		n = 1
	}
	return
}

func toEncodedPath(nibs []byte, isLeaf bool) []byte {
	isLeafVar := uint8(boolToInt(isLeaf))
	prefix := isLeafVar * 0x20
	if len(nibs)%2 != 0 {
		prefix += 0x10 + nibs[0]
		nibs = nibs[1:]
	}
	res := make([]byte, 0, len(nibs)%2+1)
	res = append(res, prefix)
	for c := range slices.Chunk(nibs, 2) {
		res = append(res, c[0]<<4+c[1])
	}
	return res
}

func toNibs(key []byte) []byte {
	res := make([]byte, 0, len(key)*2)
	for _, b := range key {
		res = append(res, b>>4, b&0xf)
	}
	return res
}

func prefixNibs(prefix []byte) []byte {
	if len(prefix) == 0 {
		panic("prefix cannot be empty")
	}
	ext, tail := prefix[0], prefix[1:]

	isOdd := ext&(1<<4) != 0

	isOddVar := boolToInt(isOdd)
	res := make([]byte, 0, len(tail)*2+isOddVar)
	if isOdd {
		res = append(res, ext&0xf)
	}
	for _, nib := range tail {
		res = append(res, nib>>4, nib&0xf)
	}
	return res
}

type mptNodeRef interface {
	encodeRLP(w rlp.EncoderBuffer) error
	hash() common.Hash
}

type bytesMptNodeRef []byte

func (b bytesMptNodeRef) encodeRLP(w rlp.EncoderBuffer) error {
	_, err := w.Write(b)
	return err
}

func (b bytesMptNodeRef) hash() common.Hash {
	return keccak.Keccak(b)
}

type digestMptNodeRef common.Hash

func (d digestMptNodeRef) encodeRLP(w rlp.EncoderBuffer) error {
	w.WriteBytes(d[:])
	return nil
}

func (d digestMptNodeRef) hash() common.Hash {
	return common.Hash(d)
}

type (
	nullNode   struct{}
	branchNode [16]*MptNode
	leafNode   struct {
		prefix []byte
		value  []byte
	}
	extensionNode struct {
		prefix []byte
		child  *MptNode
	}
	digestNode common.Hash
)

// marker interface, only for type checking
type mptNodeData interface {
	sealed()
}

func (*nullNode) sealed()      {}
func (*branchNode) sealed()    {}
func (*leafNode) sealed()      {}
func (*extensionNode) sealed() {}
func (*digestNode) sealed()    {}
