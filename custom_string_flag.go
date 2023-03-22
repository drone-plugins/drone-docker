package docker

import (
	"strings"
)

// CustomStringFlag is like a regular StringSlice flag but with
// semicolon as a delimiter
type CustomStringFlag struct {
	Value []string
}

func (f *CustomStringFlag) GetValue() []string {
	if f.Value == nil {
		return make([]string, 0)
	}
	return f.Value
}

func (f *CustomStringFlag) String() string {
	if f.Value == nil {
		return ""
	}
	return strings.Join(f.Value, ";")
}

func (f *CustomStringFlag) Set(v string) error {
	for _, s := range strings.Split(v, ";") {
		s = strings.TrimSpace(s)
		f.Value = append(f.Value, s)
	}
	return nil
}
