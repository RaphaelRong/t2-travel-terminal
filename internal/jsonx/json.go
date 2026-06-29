package jsonx

import "encoding/json"

// Object returns a non-empty JSON object when raw is empty or invalid.
func Object(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage(`{}`)
	}
	return raw
}

// MustMarshalObject marshals value and falls back to an empty JSON object.
func MustMarshalObject(value any) json.RawMessage {
	body, err := json.Marshal(value)
	if err != nil || len(body) == 0 {
		return json.RawMessage(`{}`)
	}
	return body
}

// StringMap decodes an object into a string map. Invalid values return nil.
func StringMap(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var result map[string]string
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return result
}
