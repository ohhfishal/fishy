package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringArray is a custom type for storing string slices as JSON in SQLite
type StringArray []string

// Value implements driver.Valuer - converts StringArray to database value
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan implements sql.Scanner - converts database value to StringArray
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for StringArray: %T", value)
	}

	return json.Unmarshal(bytes, sa)
}
