package builder

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type PatchOption string

func (p PatchOption) Enabled() bool {
	value := strings.TrimSpace(string(p))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "false", "no", "none":
		return false
	default:
		return true
	}
}

func (p PatchOption) String() string {
	return string(p)
}

func (p *PatchOption) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*p = ""
		return nil
	}
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("patch must be a scalar value")
	}
	switch value.Tag {
	case "!!bool":
		if strings.EqualFold(value.Value, "true") {
			*p = PatchOption("true")
		} else {
			*p = ""
		}
		return nil
	case "!!null":
		*p = ""
		return nil
	default:
		*p = PatchOption(value.Value)
		return nil
	}
}
