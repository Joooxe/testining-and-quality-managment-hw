package contract

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Shape string

const (
	Order Shape = "order"
	List  Shape = "list"
	Error Shape = "error"
)

const schemaURL = "urn:pizza-delivery:responses"

//go:embed responses.json
var schemaJSON []byte

type Schemas map[Shape]*jsonschema.Schema

func Load() (Schemas, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		return nil, fmt.Errorf("read response schemas: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource(schemaURL, doc); err != nil {
		return nil, err
	}

	schemas := make(Schemas)
	for _, shape := range []Shape{Order, List, Error} {
		schema, err := compiler.Compile(schemaURL + "#/$defs/" + string(shape))
		if err != nil {
			return nil, fmt.Errorf("compile %s schema: %w", shape, err)
		}
		schemas[shape] = schema
	}
	return schemas, nil
}

func (s Schemas) Validate(shape Shape, body []byte) error {
	schema, ok := s[shape]
	if !ok {
		return fmt.Errorf("unknown response schema %q", shape)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	return schema.Validate(instance)
}
