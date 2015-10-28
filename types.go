package main

import "encoding/json"

// StrSlice representes a string or an array of strings.
// We need to override the json decoder to accept both options.
type StrSlice struct {
	parts []string
}

// UnmarshalJSON decodes the byte slice whether it's a string or an array of strings.
// This method is needed to implement json.Unmarshaler.
func (e *StrSlice) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	p := make([]string, 0, 1)
	if err := json.Unmarshal(b, &p); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		p = append(p, s)
	}

	e.parts = p
	return nil
}

// Len returns the number of parts of the StrSlice.
func (e *StrSlice) Len() int {
	if e == nil {
		return 0
	}
	return len(e.parts)
}

// Slice gets the parts of the StrSlice as a Slice of string.
func (e *StrSlice) Slice() []string {
	if e == nil {
		return nil
	}
	return e.parts
}
