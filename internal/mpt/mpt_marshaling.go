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
			var data BranchNode
			if err := json.Unmarshal(val, &data); err != nil {
				return err
			}
			m.Data = &data
		case "Leaf":
			var data [2]json.RawMessage
			if err := json.Unmarshal(val, &data); err != nil {
				return err
			}
			m.Data = &LeafNode{
				Prefix: data[0],
				Value:  data[1],
			}
		case "Extension":
			var data [2]json.RawMessage
			if err := json.Unmarshal(val, &data); err != nil {
				return err
			}
			var child MptNode
			if err := json.Unmarshal(data[1], &child); err != nil {
				return err
			}
			ext := &ExtensionNode{
				Prefix: data[0],
				Child:  &child,
			}
			m.Data = ext
		case "Digest":
			var data DigestNode
			if err := json.Unmarshal(val, &data); err != nil {
				return err
			}
			m.Data = &data
		case "Null":
			m.Data = &NullNode{}
		default:
			return fmt.Errorf("unknown MptNodeData type: %s", key)
		}
	}
	return nil
}
