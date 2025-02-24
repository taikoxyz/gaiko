package mpt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/taikoxyz/gaiko/internal"
)

type MptNode struct {
	data      mptNodeData
	cachedRef mptNodeRef
}

var _ rlp.Encoder = (*MptNode)(nil)

func newMptNode(data mptNodeData) *MptNode {
	return &MptNode{
		data: data,
	}
}

func New() *MptNode {
	return newMptNode(&nullNode{})
}

func (m *MptNode) EncodeRLP(_w io.Writer) error {
	w := rlp.NewEncoderBuffer(_w)
	switch data := m.data.(type) {
	case *nullNode:
		w.Write(rlp.EmptyString)
	case *branchNode:
		_tmp0 := w.List()
		for _, child := range data {
			if child == nil {
				w.Write(rlp.EmptyString)
			} else {
				if err := child.refEncode(w); err != nil {
					return err
				}
			}
		}
		w.Write(rlp.EmptyString)
		w.ListEnd(_tmp0)
	case *leafNode:
		_tmp0 := w.List()
		w.WriteBytes(data.prefix)
		w.WriteBytes(data.value)
		w.ListEnd(_tmp0)
	case *extensionNode:
		_tmp0 := w.List()
		w.WriteBytes(data.prefix)
		if err := data.child.refEncode(w); err != nil {
			return err
		}
		w.ListEnd(_tmp0)
	case *digestNode:
		w.WriteBytes(data[:])
	default:
		return fmt.Errorf("unknown MptNodeData type: %T", data)
	}
	return w.Flush()
}

func (m *MptNode) Clear() {
	m.data = &nullNode{}
	m.cachedRef = nil
}

func (m *MptNode) Hash() (common.Hash, error) {
	switch m.data.(type) {
	case *nullNode:
		return types.EmptyRootHash, nil
	default:
		ref, err := m.ref()
		if err != nil {
			return common.Hash{}, err
		}
		return ref.hash(), nil
	}
}

func (m *MptNode) IsEmpty() bool {
	switch m.data.(type) {
	case *nullNode:
		return true
	default:
		return false
	}
}

func (m *MptNode) IsDigest() bool {
	switch m.data.(type) {
	case *digestNode:
		return true
	default:
		return false
	}
}

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

func (m *MptNode) Get(key []byte) ([]byte, error) {
	return m.get(toNibs(key))
}

func (m *MptNode) Delete(key []byte) (bool, error) {
	return m.delete(toNibs(key))
}

func (m *MptNode) Insert(key []byte, value []byte) (bool, error) {
	return m.insert(toNibs(key), value)
}

func (m *MptNode) InsertRLP(key []byte, value interface{}) (bool, error) {
	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		return false, err
	}
	return m.insert(toNibs(key), data)
}

func (m *MptNode) insert(keyNibs []byte, value []byte) (bool, error) {
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
		if child == nil {
			data[idx] = newMptNode(&leafNode{
				prefix: toEncodedPath(tail, true),
				value:  value,
			})
		} else {
			ok, err := child.insert(tail, value)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
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
			data.child = New()
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
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
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
		i, tail := keyNibs[0], keyNibs[1:]
		child := data[i]
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
			data[i] = nil
		}
		nextIdx := 0
		remaining := 0
		var nextChild *MptNode
		for idx, child := range data {
			if child != nil {
				remaining += 1
				if nextChild == nil {
					nextIdx = idx
					nextChild = child
				}
			}
		}
		if remaining == 1 {
			switch data := nextChild.data.(type) {
			case *leafNode:
				newNibs := append([]byte{uint8(nextIdx)}, prefixNibs(data.prefix)...)
				m.data = &leafNode{
					prefix: toEncodedPath(newNibs, true),
					value:  data.value,
				}
			case *extensionNode:
				newNibs := append([]byte{uint8(nextIdx)}, prefixNibs(data.prefix)...)
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
		if bytes.HasPrefix(keyNibs, selfNibs) {
			ok, err := data.child.delete(bytes.TrimPrefix(keyNibs, selfNibs))
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
			selfNibs := append(selfNibs, prefixNibs(data.prefix)...)
			m.data = &leafNode{
				prefix: toEncodedPath(selfNibs, true),
				value:  data.value,
			}
		case *extensionNode:
			selfNibs := append(selfNibs, prefixNibs(data.prefix)...)
			m.data = &extensionNode{
				prefix: toEncodedPath(selfNibs, true),
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
		idx := keyNibs[0]
		tail := keyNibs[1:]

		if data[idx] == nil {
			return nil, nil
		}
		return data[idx].get(tail)
	case *leafNode:
		if len(keyNibs) == 0 {
			return data.value, nil
		}
		return nil, nil
	case *extensionNode:
		if len(keyNibs) == 0 {
			return nil, nil
		}
		prefix := prefixNibs(data.prefix)
		if bytes.HasPrefix(keyNibs, prefix) {
			data.child.get(bytes.TrimPrefix(keyNibs, prefix))
		}
		return nil, nil
	case *digestNode:
		return nil, fmt.Errorf("node not resolved: %s", data)
	}
	return nil, nil
}

func toEncodedPath(nibs []byte, isLeaf bool) []byte {
	isLeafVar := uint8(0)
	if isLeaf {
		isLeafVar = 1
	}

	prefix := isLeafVar * 0x20
	if len(nibs)%2 != 0 {
		prefix += 0x10 + nibs[0]
		nibs = nibs[1:]
	}
	res := make([]byte, 0, len(nibs)%2+1)
	res = append(res, prefix)
	for i := 0; i < len(nibs); i += 2 {
		res = append(res, nibs[i]<<4+nibs[i+1])
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

	isOaddVar := 0
	if isOdd {
		isOaddVar = 1
	}
	res := make([]byte, 0, len(tail)*2+isOaddVar)
	if isOdd {
		res = append(res, ext&0xf)
	}
	for _, nib := range tail {
		res = append(res, nib>>4, nib&0xf)
	}
	return res
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
				m.cachedRef = digestMptNodeRef(common.BytesToHash(internal.Keccak(encoded)))
			}
		}
	}
	return m.cachedRef, nil
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
	return common.BytesToHash(internal.Keccak(b))
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
	_private()
}

func (*nullNode) _private()      {}
func (*branchNode) _private()    {}
func (*leafNode) _private()      {}
func (*extensionNode) _private() {}
func (*digestNode) _private()    {}
