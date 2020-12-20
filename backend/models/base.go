package models

import (
	"database/sql"
	"encoding/json"
)

type NullString struct {
	sql.NullString
}
type NullInt32 struct {
	sql.NullInt32
}

func (t NullString) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.String)
}
func (t *NullString) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		t.String, t.Valid = "", false
		return nil
	}
	t.String, t.Valid = str[1:len(str)-1], true
	return nil
}

func (t *NullInt32) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.Int32)
}

func (t *NullInt32) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &t.Int32)
	t.Valid = err == nil
	return nil
}
