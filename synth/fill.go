package synth

import (
	"encoding/json"
	"slices"
)

// Fill returns a copy of p in which the named track ("kick", "snare" or
// "hat") plays on every step from from to to, both included. Steps are
// numbered 1 to Steps. Steps the track already played outside the span are
// kept. p is not modified and the result shares no memory with it. Every
// error wraps ErrInvalid and comes with the zero Pattern.
func Fill(p Pattern, track string, from, to int) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	if from < 1 || to > Steps || from > to {
		return Pattern{}, invalid("span %d to %d must satisfy 1 <= from <= to <= %d", from, to, Steps)
	}
	out := p
	out.ID = cloneRaw(p.ID)
	out.Votes = cloneRaw(p.Votes)
	out.CreatedAt = cloneRaw(p.CreatedAt)
	out.Tracks.Kick = slices.Clone(p.Tracks.Kick)
	out.Tracks.Snare = slices.Clone(p.Tracks.Snare)
	out.Tracks.Hat = slices.Clone(p.Tracks.Hat)
	var target *[]bool
	switch track {
	case "kick":
		target = &out.Tracks.Kick
	case "snare":
		target = &out.Tracks.Snare
	case "hat":
		target = &out.Tracks.Hat
	default:
		return Pattern{}, invalid("track must be \"kick\", \"snare\" or \"hat\"")
	}
	for i := from - 1; i <= to-1; i++ {
		(*target)[i] = true
	}
	return out, nil
}

// cloneRaw copies a raw JSON field so the result never shares its bytes.
func cloneRaw(m json.RawMessage) json.RawMessage {
	return append(json.RawMessage{}, m...)
}
