package mpt

import (
	"fmt"
	"io"

	json "github.com/goccy/go-json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

var _ json.Unmarshaler = (*MptNode)(nil)
var _ rlp.Encoder = (*MptNode)(nil)

func (m *MptNode) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, rawData := range raw {
		// try Null
		// {"data": "Null" }
		var null string
		if err := json.Unmarshal(rawData, &null); err == nil {
			if null == "Null" {
				m.data = &nullNode{}
				return nil
			}
		}
		data := map[string]json.RawMessage{}
		if err := json.Unmarshal(rawData, &data); err != nil {
			return err
		}
		for key, val := range data {
			switch key {
			case "Branch":
				// {"data": {"Branch": [{...}]}}
				var data branchNode
				if err := json.Unmarshal(val, &data); err != nil {
					return err
				}
				m.data = &data
			case "Leaf":
				// {"data": {"Leaf": [prefix, value]}}
				var data [2][]byte
				if err := json.Unmarshal(val, &data); err != nil {
					return err
				}
				m.data = &leafNode{
					prefix: data[0],
					value:  data[1],
				}
			case "Extension":
				// {"data": {"Extension": [prefix, child]}}
				var data [2]json.RawMessage
				if err := json.Unmarshal(val, &data); err != nil {
					return err
				}
				var prefix []byte
				if err := json.Unmarshal(data[0], &prefix); err != nil {
					return err
				}
				var child MptNode
				if err := json.Unmarshal(data[1], &child); err != nil {
					return err
				}
				ext := &extensionNode{
					prefix: prefix,
					child:  &child,
				}
				m.data = ext
			case "Digest":
				// {"data": {"Digest": ""}}
				var data common.Hash
				if err := json.Unmarshal(val, &data); err != nil {
					return err
				}
				m.data = (*digestNode)(&data)
			default:
				return fmt.Errorf("unknown MptNodeData type: %s", key)
			}
		}
	}
	return nil
}

func (m *MptNode) EncodeRLP(_w io.Writer) error {
	return m.encodeRLP(_w, nil)
}

func (m *MptNode) encodeRLP(_w io.Writer, wlk func([]byte)) error {
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
				if err := child.refEncodeRLP(w, wlk); err != nil {
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
		if err := data.child.refEncodeRLP(w, wlk); err != nil {
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

type walkMptNode struct {
	mpt    *MptNode
	walker func([]byte)
}

func (m *walkMptNode) EncodeRLP(_w io.Writer) error {
	return m.mpt.encodeRLP(_w, m.walker)
}
