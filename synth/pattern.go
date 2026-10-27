// Package synth is a pure, deterministic drum synth: it validates a 16-step
// pattern, renders it to 16-bit mono PCM, encodes WAV and detects onsets. It
// has no HTTP, clock or global randomness, so its output is bit-for-bit
// reproducible.
package synth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Steps is the number of steps in a pattern.
const Steps = 16

// Tracks holds one row of steps per instrument. Unknown instruments are
// rejected when decoding with DisallowUnknownFields.
type Tracks struct {
	Kick  []bool `json:"kick"`
	Snare []bool `json:"snare"`
	Hat   []bool `json:"hat"`
}

// Pattern is the shared pattern JSON. ID, Votes and CreatedAt are
// server-assigned elsewhere: they are accepted and ignored here.
type Pattern struct {
	ID        json.RawMessage `json:"id"`
	Name      string          `json:"name"`
	Author    string          `json:"author"`
	BPM       int             `json:"bpm"`
	Swing     int             `json:"swing"`
	Tracks    Tracks          `json:"tracks"`
	Votes     json.RawMessage `json:"votes"`
	CreatedAt json.RawMessage `json:"createdAt"`
}

// ErrInvalid wraps every validation failure.
var ErrInvalid = errors.New("invalid pattern")

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// Validate applies the shared validation rules.
func (p Pattern) Validate() error {
	if err := validateName(p.Name); err != nil {
		return err
	}
	if err := validateAuthor(p.Author); err != nil {
		return err
	}
	if p.BPM < 60 || p.BPM > 200 {
		return invalid("bpm must be an integer from 60 to 200")
	}
	if p.Swing < 0 || p.Swing > 60 {
		return invalid("swing must be an integer from 0 to 60")
	}
	for _, t := range []struct {
		name  string
		steps []bool
	}{{"kick", p.Tracks.Kick}, {"snare", p.Tracks.Snare}, {"hat", p.Tracks.Hat}} {
		if len(t.steps) != Steps {
			return invalid("track %q must have exactly %d steps", t.name, Steps)
		}
	}
	return nil
}

func validateName(name string) error {
	name = strings.TrimSpace(name)
	n := utf8.RuneCountInString(name)
	if n < 1 || n > 40 {
		return invalid("name must be 1 to 40 characters")
	}
	if strings.ContainsFunc(name, unicode.IsControl) {
		return invalid("name must not contain control characters")
	}
	return nil
}

func validateAuthor(author string) error {
	if len(author) < 1 || len(author) > 24 {
		return invalid("author must be 1 to 24 characters")
	}
	for _, c := range author {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
		if !ok {
			return invalid("author may only contain letters, digits, _ and -")
		}
	}
	return nil
}
