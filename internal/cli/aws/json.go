package aws

import (
	"encoding/json"
	"io"
)

// JSONEncoder wraps json.Encoder with pretty printing
type JSONEncoder struct {
	enc *json.Encoder
}

// NewJSONEncoder creates a new JSON encoder
func NewJSONEncoder(w io.Writer) *JSONEncoder {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return &JSONEncoder{enc: enc}
}

// Encode encodes a value to JSON
func (e *JSONEncoder) Encode(v any) error {
	return e.enc.Encode(v)
}

// SetIndent sets the indentation for the encoder
func (e *JSONEncoder) SetIndent(prefix, indent string) {
	e.enc.SetIndent(prefix, indent)
}
