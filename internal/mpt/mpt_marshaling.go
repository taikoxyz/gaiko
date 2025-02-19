package mpt

import (
	"encoding/json"
	"fmt"
)

var _ json.Unmarshaler = (*MptNode)(nil)

func (m *MptNode) UnmarshalJSON(data []byte) error {
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
			m.Data = &inner
		case "Leaf":
			var inner [2]json.RawMessage
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			m.Data = &LeafNode{
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
			m.Data = ext
		case "Digest":
			var inner DigestNode
			if err := json.Unmarshal(val, &inner); err != nil {
				return err
			}
			m.Data = &inner
		case "Null":
			m.Data = &NullNode{}
		default:
			return fmt.Errorf("unknown MptNodeData type: %s", key)
		}
	}
	return nil
}
