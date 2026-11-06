package synth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// maxParseBytes is the largest input Parse accepts.
const maxParseBytes = 16384

// ErrTooLong is returned by Parse when its input exceeds maxParseBytes.
var ErrTooLong = errors.New("pattern too long")

// Parse decodes one pattern from JSON bytes, rejecting unknown fields and
// trailing data, then validates it with Validate. Input over 16384 bytes is
// refused before decoding with ErrTooLong; any other failure, including a
// decoded-but-invalid pattern, is reported through ErrInvalid.
func Parse(data []byte) (Pattern, error) {
	var p Pattern
	if len(data) > maxParseBytes {
		return p, fmt.Errorf("%w: body exceeds %d bytes", ErrTooLong, maxParseBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return p, fmt.Errorf("%w: body is not a valid pattern JSON object: %s", ErrInvalid, err.Error())
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return p, fmt.Errorf("%w: unexpected data after the JSON object", ErrInvalid)
	}
	if err := p.Validate(); err != nil {
		return p, err
	}
	return p, nil
}
