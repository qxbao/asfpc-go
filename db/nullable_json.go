package db

import (
	"database/sql/driver"
	"encoding/json"
)

type NullableJSON []byte

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
			// Copy the bytes to ensure we get the full data
			*nj = make([]byte, len(v))
			copy(*nj, v)
		}
		return nil
	case string:
		if v == "" {
			*nj = []byte("[]")
		} else {
			// Don't validate, PostgreSQL jsonb is already valid
			*nj = []byte(v)
		}
		return nil
	default:
		*nj = []byte("[]")
		return nil
	}
}

func (nj NullableJSON) Value() (driver.Value, error) {
	if len(nj) == 0 {
		return []byte("[]"), nil
	}
	return []byte(nj), nil
}

func (nj NullableJSON) MarshalJSON() ([]byte, error) {
	if len(nj) == 0 {
		return []byte("[]"), nil
	}
	// Don't validate, PostgreSQL jsonb is already valid
	return []byte(nj), nil
}

func (nj *NullableJSON) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*nj = []byte("[]")
		return nil
	}
	*nj = data
	return nil
}

func (nj NullableJSON) ToRawMessage() json.RawMessage {
	if len(nj) == 0 {
		return json.RawMessage("[]")
	}
	return json.RawMessage(nj)
}
