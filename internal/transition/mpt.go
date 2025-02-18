package transition

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const nilKindString = 0x80

func writeNilString(w rlp.EncoderBuffer) {
	w.Write([]byte{nilKindString})
}

type MptNode struct {
	Data      *MptNodeData `json:"data"`
	cachedRef MptNodeRef   `json:"-"`
}

func NewMptNode(inner interface{}) *MptNode {
	return &MptNode{
		Data: NewMptNodeData(inner),
	}
}

func (m *MptNode) EncodeRLP(_w io.Writer) error {
	w := rlp.NewEncoderBuffer(_w)
	switch inner := m.Data.inner.(type) {
	case *NullNode:
		writeNilString(w)
	case *DigestNode:
		w.WriteBytes(inner[:])
	case *BranchNode:
		idx := w.List()
		for _, child := range inner {
			if child == nil {
				writeNilString(w)
			} else {
				if err := child.refEncode(w); err != nil {
					return err
				}
			}
		}
		w.ListEnd(idx)
		writeNilString(w)
	case *LeafNode:
		idx := w.List()
		w.WriteBytes(inner.Prefix)
		w.WriteBytes(inner.Value)
		w.ListEnd(idx)
	case *ExtensionNode:
		idx := w.List()
		w.WriteBytes(inner.Prefix)
		if err := inner.Child.refEncode(w); err != nil {
			return err
		}
		w.ListEnd(idx)
	default:
		return fmt.Errorf("unknown MptNodeData type: %T", inner)
	}
	return nil
}

func (m *MptNode) Clear() {
	m.Data = NewMptNodeData(NullNode{})
	m.cachedRef = nil
}

func (m *MptNode) Hash() (common.Hash, error) {
	switch m.Data.inner.(type) {
	case *NullNode:
		return types.EmptyRootHash, nil
	default:
		ref, err := m.ref()
		if err != nil {
			return common.Hash{}, err
		}
		return ref.Hash(), nil
	}
}

func (m *MptNode) IsEmpty() bool {
	switch m.Data.inner.(type) {
	case *NullNode:
		return true
	default:
		return false
	}
}

func (m *MptNode) Nibs() []byte {
	switch inner := m.Data.inner.(type) {
	case *NullNode, *BranchNode, *DigestNode:
		return nil
	case *LeafNode:
		return prefixNibs(inner.Prefix)
	case *ExtensionNode:
		return prefixNibs(inner.Prefix)
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

func (m *MptNode) insert(keyNibs []byte, value []byte) (bool, error) {
	switch inner := m.Data.inner.(type) {
	case *NullNode:
		m.Data = NewMptNodeData(&LeafNode{
			Prefix: toEncodedPath(keyNibs, true),
			Value:  value,
		})
	case *BranchNode:
		if len(keyNibs) == 0 {
			return false, errors.New("branch node with value")
		}
		idx, tail := keyNibs[0], keyNibs[1:]
		child := inner[idx]
		if child == nil {
			inner[idx] = NewMptNode(&LeafNode{
				Prefix: toEncodedPath(tail, true),
				Value:  value,
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
	case *LeafNode:
		selfNibs := prefixNibs(inner.Prefix)
		commonLen := lcp(selfNibs, keyNibs)
		if commonLen == len(selfNibs) && commonLen == len(keyNibs) {
			if slices.Equal(inner.Value, value) {
				return false, nil
			}
			inner.Value = value
		} else if commonLen == len(selfNibs) || commonLen == len(keyNibs) {
			return false, errors.New("branch node with value")
		} else {
			splitPoint := commonLen + 1
			branch := &BranchNode{}
			branch[selfNibs[commonLen]] = NewMptNode(&LeafNode{
				Prefix: toEncodedPath(selfNibs[splitPoint:], true),
				Value:  inner.Value,
			})
			branch[keyNibs[commonLen]] = NewMptNode(&LeafNode{
				Prefix: toEncodedPath(keyNibs[splitPoint:], true),
				Value:  value,
			})
			if commonLen > 0 {
				m.Data = NewMptNodeData(&ExtensionNode{
					Prefix: toEncodedPath(selfNibs[:commonLen], false),
					Child:  NewMptNode(branch),
				})
			} else {
				m.Data = NewMptNodeData(branch)
			}
		}
	case *ExtensionNode:
		selfNibs := prefixNibs(inner.Prefix)
		commonLen := lcp(selfNibs, keyNibs)
		if commonLen == len(selfNibs) {
			ok, err := inner.Child.insert(keyNibs[commonLen:], value)
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
			branch := &BranchNode{}
			if splitPoint < len(selfNibs) {
				branch[selfNibs[commonLen]] = NewMptNode(&ExtensionNode{
					Prefix: toEncodedPath(selfNibs[splitPoint:], false),
					Child:  inner.Child,
				})
			} else {
				branch[selfNibs[commonLen]] = inner.Child
			}
			branch[keyNibs[commonLen]] = NewMptNode(&LeafNode{
				Prefix: toEncodedPath(keyNibs[splitPoint:], true),
				Value:  value,
			})
			if commonLen > 0 {
				m.Data = NewMptNodeData(&ExtensionNode{
					Prefix: toEncodedPath(selfNibs[:commonLen], false),
					Child:  NewMptNode(branch),
				})
			} else {
				m.Data = NewMptNodeData(branch)
			}
		}
	case *DigestNode:
		return false, fmt.Errorf("node not resolved: %s", inner)
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
	switch inner := m.Data.inner.(type) {
	case *NullNode:
		return false, nil
	case *BranchNode:
		if len(keyNibs) == 0 {
			return false, errors.New("branch node with value")
		}
		i, tail := keyNibs[0], keyNibs[1:]
		child := inner[i]
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
			inner[i] = nil
		}
		nextIdx := 0
		remaining := 0
		var nextChild *MptNode
		for idx, child := range inner {
			if child != nil {
				remaining += 1
				if nextChild == nil {
					nextIdx = idx
					nextChild = child
				}
				break
			}
		}
		if remaining == 1 {
			switch inner := nextChild.Data.inner.(type) {
			case *LeafNode:
				newNibs := append([]byte{uint8(nextIdx)}, prefixNibs(inner.Prefix)...)
				m.Data = NewMptNodeData(&LeafNode{
					Prefix: toEncodedPath(newNibs, true),
					Value:  inner.Value,
				})
			case *ExtensionNode:
				newNibs := append([]byte{uint8(nextIdx)}, prefixNibs(inner.Prefix)...)
				m.Data = NewMptNodeData(&ExtensionNode{
					Prefix: toEncodedPath(newNibs, false),
					Child:  inner.Child,
				})
			case *BranchNode, *DigestNode:
				m.Data = NewMptNodeData(&ExtensionNode{
					Prefix: toEncodedPath([]byte{byte(nextIdx)}, false),
					Child:  nextChild,
				})
			case *NullNode:
				panic("unreachable")
			}
		}
	case *LeafNode:
		if !slices.Equal(prefixNibs(inner.Prefix), keyNibs) {
			return false, nil
		}
		m.Data = NewMptNodeData(&NullNode{})
	case *ExtensionNode:
		selfNibs := prefixNibs(inner.Prefix)
		if bytes.HasPrefix(keyNibs, selfNibs) {
			ok, err := inner.Child.delete(bytes.TrimPrefix(keyNibs, selfNibs))
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		} else {
			return false, nil
		}

		switch inner := inner.Child.Data.inner.(type) {
		case *NullNode:
			m.Data = NewMptNodeData(&NullNode{})
		case *LeafNode:
			selfNibs := append(selfNibs, prefixNibs(inner.Prefix)...)
			m.Data = NewMptNodeData(&LeafNode{
				Prefix: toEncodedPath(selfNibs, true),
				Value:  inner.Value,
			})
		case *ExtensionNode:
			selfNibs := append(selfNibs, prefixNibs(inner.Prefix)...)
			m.Data = NewMptNodeData(&ExtensionNode{
				Prefix: toEncodedPath(selfNibs, true),
				Child:  inner.Child,
			})
		case *BranchNode, *DigestNode:
		}
	case *DigestNode:
		return false, fmt.Errorf("node not resolved: %s", inner)
	}
	m.cachedRef = nil
	return true, nil
}

func (m *MptNode) get(keyNibs []byte) ([]byte, error) {
	switch inner := m.Data.inner.(type) {
	case *NullNode:
		return nil, nil
	case *BranchNode:
		if len(keyNibs) == 0 {
			return nil, nil
		}
		idx := keyNibs[0]
		tail := keyNibs[1:]

		if inner[idx] == nil {
			return nil, nil
		}
		return inner[idx].get(tail)
	case *LeafNode:
		if len(keyNibs) == 0 {
			return inner.Value, nil
		}
		return nil, nil
	case *ExtensionNode:
		if len(keyNibs) == 0 {
			return nil, nil
		}
		prefix := prefixNibs(inner.Prefix)
		if bytes.HasPrefix(keyNibs, prefix) {
			inner.Child.get(bytes.TrimPrefix(keyNibs, prefix))
		}
		return nil, nil
	case *DigestNode:
		return nil, fmt.Errorf("node not resolved: %s", inner)
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
	if prefix == nil || len(prefix) == 0 {
		return nil
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
	ref.EncodeRLP(w)
	return nil
}

func (m *MptNode) ref() (MptNodeRef, error) {
	if m.cachedRef == nil {
		switch inner := m.Data.inner.(type) {
		case *NullNode:
			m.cachedRef = BytesMptNodeRef{0x80}
		case *DigestNode:
			m.cachedRef = DigestMptNodeRef(*inner)
		default:
			encoded, err := rlp.EncodeToBytes(m)
			if err != nil {
				return nil, err
			}
			if len(encoded) < 32 {
				m.cachedRef = BytesMptNodeRef(encoded)
			} else {
				m.cachedRef = DigestMptNodeRef(common.BytesToHash(keccak(encoded)))
			}
		}
	}
	return m.cachedRef, nil
}

type MptNodeRef interface {
	EncodeRLP(w rlp.EncoderBuffer)
	Hash() common.Hash
	Len() int
}

type BytesMptNodeRef []byte

func (b BytesMptNodeRef) EncodeRLP(w rlp.EncoderBuffer) {
	w.WriteBytes(b)
}

func (b BytesMptNodeRef) Hash() common.Hash {
	return common.BytesToHash(keccak(b))
}

func (b BytesMptNodeRef) Len() int {
	return len(b)
}

type DigestMptNodeRef common.Hash

func (d DigestMptNodeRef) EncodeRLP(w rlp.EncoderBuffer) {
	w.Write([]byte{nilKindString + 32})
	w.WriteBytes(d[:])
}

func (d DigestMptNodeRef) Hash() common.Hash {
	return common.Hash(d)
}

func (d DigestMptNodeRef) Len() int {
	return 33
}

type MptNodeData struct {
	inner interface{}
}

func NewMptNodeData(inner interface{}) *MptNodeData {
	return &MptNodeData{inner: inner}
}

type (
	NullNode   struct{}
	BranchNode [16]*MptNode
	LeafNode   struct {
		Prefix []byte
		Value  []byte
	}
	ExtensionNode struct {
		Prefix []byte
		Child  *MptNode
	}
	DigestNode common.Hash
)

func (m *MptNodeData) Unmarshal(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for key, val := range raw {
		switch key {
		case "Branch":
			var inner BranchNode
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			m.inner = &inner
		case "Leaf":
			var inner [2]json.RawMessage
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			m.inner = &LeafNode{
				Prefix: inner[0],
				Value:  inner[1],
			}
		case "Extension":
			var inner [2]json.RawMessage
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			var child MptNode
			if err := json.Unmarshal(inner[1], &child); err != nil {
				return err
			}
			ext := &ExtensionNode{
				Prefix: inner[0],
				Child:  &child,
			}
			m.inner = ext
		case "Digest":
			var inner DigestNode
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			m.inner = &inner
		case "Null":
			m.inner = &NullNode{}
		default:
			return fmt.Errorf("unknown MptNodeData type: %s", key)
		}
	}
	return nil
}
