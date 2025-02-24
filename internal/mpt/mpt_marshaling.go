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
				var data [2]json.RawMessage
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
				var child MptNode
				if err := json.Unmarshal(data[1], &child); err != nil {
					return err
				}
				ext := &extensionNode{
					prefix: data[0],
					child:  &child,
				}
				m.data = ext
			case "Digest":
				// {"data": {"Digest": ""}}
				var data digestNode
				if err := json.Unmarshal(val, &data); err != nil {
					return err
				}
				m.data = &data
			default:
				return fmt.Errorf("unknown MptNodeData type: %s", key)
			}
		}
	}
	return nil
}
