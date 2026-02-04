package insomnia

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidYAML       = errors.New("invalid YAML format")
	ErrMissingType       = errors.New("missing or invalid type field")
	ErrMissingCollection = errors.New("missing or empty collection field")
)

func Parse(content []byte) (*InsomniaExport, error) {
	var export InsomniaExport
	if err := yaml.Unmarshal(content, &export); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidYAML, err)
	}

	if export.Type == "" {
		return nil, ErrMissingType
	}

	if export.Type != "collection.insomnia.rest/5.0" {
		return nil, fmt.Errorf("%w: expected 'collection.insomnia.rest/5.0', got '%s'", ErrInvalidYAML, export.Type)
	}

	if len(export.Collection) == 0 {
		return nil, ErrMissingCollection
	}

	return &export, nil
}
