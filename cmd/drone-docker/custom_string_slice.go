package main

import (
	"strings"
)

// CustomStringSliceFlag is like a regular StringSlice flag but with
// semicolon as a delimiter
type CustomStringSliceFlag struct {
	Value []string
}

func (f *CustomStringSliceFlag) GetValue() []string {
	if f.Value == nil {
		return make([]string, 0)
	}
	return f.Value
}

func (f *CustomStringSliceFlag) String() string {
	if f.Value == nil {
		return ""
	}
	return strings.Join(f.Value, ";")
}

func (f *CustomStringSliceFlag) Set(v string) error {
	for _, s := range strings.Split(v, ";") {
		s = strings.TrimSpace(s)
		f.Value = append(f.Value, s)
	}
	return nil
}
