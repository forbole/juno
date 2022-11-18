package config

import (
	"gopkg.in/yaml.v3"

	"github.com/saifullah619/juno/v3/node/local"
	"github.com/saifullah619/juno/v3/node/remote"
)

const (
	TypeRemote = "remote"
	TypeLocal  = "local"
	TypeNone   = "none"
)

type Config struct {
	Type    string  `yaml:"type"`
	Details Details `yaml:"-"`
}

func NewConfig(nodeType string, details Details) Config {
	return Config{
		Type:    nodeType,
		Details: details,
	}
}

func DefaultConfig() Config {
	return NewConfig(TypeRemote, remote.DefaultDetails())
}

func (s *Config) UnmarshalYAML(n *yaml.Node) error {
	type S Config
	type T struct {
		*S      `yaml:",inline"`
		Details yaml.Node `yaml:"config"`
	}

	obj := &T{S: (*S)(s)}
	if err := n.Decode(obj); err != nil {
		return err
	}

	switch obj.Type {
	case TypeRemote:
		s.Details = new(remote.Details)
	case TypeLocal:
		s.Details = new(local.Details)
	default:
		panic("unknown node type")
	}

	return obj.Details.Decode(s.Details)
}

func (s Config) MarshalYAML() (interface{}, error) {
	type S Config
	type T struct {
		S       `yaml:",inline"`
		Details Details `yaml:"config"`
	}

	obj := &T{S: S(s)}
	obj.Details = s.Details

	return obj, nil
}

type Details interface {
	Validate() error
}
