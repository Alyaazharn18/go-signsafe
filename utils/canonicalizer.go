package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

func canonicalJSONStringify(val any) ([]byte, error) {
	var buf bytes.Buffer

	switch v := val.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if v {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case float64:
		numBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal number %v: %w", v, err)
		}
		buf.Write(numBytes)
	case string:
		strBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal string %q: %w", v, err)
		}
		buf.Write(strBytes)
	case []any:
		buf.WriteString("[")
		for i, elem := range v {
			if i > 0 {
				buf.WriteString(",")
			}
			elemBytes, err := canonicalJSONStringify(elem)
			if err != nil {
				return nil, err
			}
			buf.Write(elemBytes)
		}
		buf.WriteString("]")
	case map[string]any:
		buf.WriteString("{")
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				buf.WriteString(",")
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal object key %q: %w", k, err)
			}
			buf.Write(keyBytes)
			buf.WriteString(":")

			valBytes, err := canonicalJSONStringify(v[k])
			if err != nil {
				return nil, err
			}
			buf.Write(valBytes)
		}
		buf.WriteString("}")
	default:
		fallbackBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("unsupported or unmarshalable type %T for canonicalization: %w", v, err)
		}
		buf.Write(fallbackBytes)
	}

	return buf.Bytes(), nil
}

func CanonicalizeBody(bodyBytes []byte) ([]byte, error) {
	var parsed any
	// First, unmarshal the raw JSON bytes into a generic Go interface{}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON bytes for manual canonicalization: %w", err)
	}

	// Then, recursively canonicalize the parsed Go value
	canonicalBytes, err := canonicalJSONStringify(parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize Go value manually: %w", err)
	}

	return canonicalBytes, nil
}
