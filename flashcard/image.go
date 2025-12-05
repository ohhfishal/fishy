package flashcard

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var _ driver.Valuer = &Image{}
var _ sql.Scanner = &Image{}

type Image struct {
	Source string `json:"source"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (i Image) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (i *Image) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for Image: %T", value)
	}

	return json.Unmarshal(bytes, i)
}
