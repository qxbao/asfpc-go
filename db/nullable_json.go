package db

import (
	"database/sql/driver"
	"encoding/json"
)

// NullableJSON is a custom type that can handle NULL JSON values from the database
type NullableJSON []byte

// Scan implements the sql.Scanner interface
func (nj *NullableJSON) Scan(value interface{}) error {
	if value == nil {
		*nj = []byte("[]")
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*nj = []byte("[]")
		} else {
			// Validate that it's valid JSON
			var test interface{}
			if err := json.Unmarshal(v, &test); err != nil {
				// Invalid JSON from database, default to empty array
				*nj = []byte("[]")
			} else {
				*nj = v
			}
		}
		return nil
	case string:
		if v == "" {
			*nj = []byte("[]")
		} else {
			// Validate that it's valid JSON
			var test interface{}
			if err := json.Unmarshal([]byte(v), &test); err != nil {
				// Invalid JSON from database, default to empty array
				*nj = []byte("[]")
			} else {
				*nj = []byte(v)
			}
		}
		return nil
	default:
		*nj = []byte("[]")
		return nil
	}
}

// Value implements the driver.Valuer interface
func (nj NullableJSON) Value() (driver.Value, error) {
	if len(nj) == 0 {
		return []byte("[]"), nil
	}
	return []byte(nj), nil
}

// MarshalJSON implements json.Marshaler
// This ensures the JSON is not double-encoded
func (nj NullableJSON) MarshalJSON() ([]byte, error) {
	if nj == nil || len(nj) == 0 {
		return []byte("[]"), nil
	}

	// Validate that the data is actually valid JSON
	var test interface{}
	if err := json.Unmarshal(nj, &test); err != nil {
		// If it's not valid JSON, return an empty array instead of failing
		return []byte("[]"), nil
	}

	// Return the raw bytes directly (already valid JSON)
	return []byte(nj), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (nj *NullableJSON) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*nj = []byte("[]")
		return nil
	}
	*nj = data
	return nil
}

// ToRawMessage converts NullableJSON to json.RawMessage
func (nj NullableJSON) ToRawMessage() json.RawMessage {
	if len(nj) == 0 {
		return json.RawMessage("[]")
	}
	return json.RawMessage(nj)
}
