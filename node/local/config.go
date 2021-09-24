package local

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// Details represents the nodeconfig.Details implementation for a local node
type Details struct {
	Home string `yaml:"home"`
}

func NewDetails(home string) *Details {
	return &Details{
		Home: home,
	}
}

func DefaultDetails() *Details {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return NewDetails(path.Join(home, ".simd"))
}

// Validate implements nodeconfig.Details
func (d *Details) Validate() error {
	if strings.TrimSpace(d.Home) == "" {
		return fmt.Errorf("home path cannot be empty")
	}

	return nil
}
