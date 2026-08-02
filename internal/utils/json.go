package utils

import (
	"encoding/json"
	"fmt"
)

// MapToJSON converts map to JSON string.
func MapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// JSONToMap parses JSON string to map.
func JSONToMap(str string) map[string]string {
	res := make(map[string]string)
	if str == "" {
		return res
	}
	_ = json.Unmarshal([]byte(str), &res)
	return res
}

// SliceToJSON converts string slice to JSON string.
func SliceToJSON(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	bytes, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

// JSONToSlice parses JSON string to slice.
func JSONToSlice(str string) []string {
	var res []string
	if str == "" {
		return res
	}
	_ = json.Unmarshal([]byte(str), &res)
	return res
}

// PrettyJSON formats raw JSON string with indentation.
func PrettyJSON(str string) string {
	var val interface{}
	if err := json.Unmarshal([]byte(str), &val); err != nil {
		return str
	}
	formatted, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return str
	}
	return string(formatted)
}

// ValidateJSON checks if input is valid JSON syntax.
func ValidateJSON(str string) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(str), &js); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	return nil
}
