package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func MapToText(m any) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("cannot convert map to json: %w", err)
	}

	return string(b), nil
}

func TextToMap[T any](s string) (T, error) {
	var m T
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return m, fmt.Errorf("cannot unmarshall string json: %w", err)
	}

	return m, nil
}

func NullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}

	return ""
}

func IntToBool(i int) bool {
	if i > 0 {
		return true
	}

	return false
}
