package server

import (
	"fmt"
	"time"
)

// requiredTime reads a tool's RFC 3339 time argument.
func requiredTime(field, value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be an RFC 3339 time, like 2026-12-31T00:00:00Z", field)
	}
	return t, nil
}

// optionalTime reads an optional RFC 3339 time argument; empty is no time.
func optionalTime(field, value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := requiredTime(field, value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// deref is a time, or the zero time for none.
func deref(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
