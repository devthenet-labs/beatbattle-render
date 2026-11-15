package synth

import (
	"bytes"
	"slices"
)

// Invert returns a copy of p in which the named track ("kick", "snare" or
// "hat") is flipped: each step plays exactly when it did not in p. p and its
// slices are not modified and the copy shares no slices with p. An unknown
// track name or a pattern Validate refuses yields the zero Pattern and an
// error wrapping ErrInvalid.
func Invert(p Pattern, track string) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	if track != "kick" && track != "snare" && track != "hat" {
		return Pattern{}, invalid("unknown track %q", track)
	}
	out := p
	out.ID = bytes.Clone(p.ID)
	out.Votes = bytes.Clone(p.Votes)
	out.CreatedAt = bytes.Clone(p.CreatedAt)
	out.Tracks = Tracks{
		Kick:  slices.Clone(p.Tracks.Kick),
		Snare: slices.Clone(p.Tracks.Snare),
		Hat:   slices.Clone(p.Tracks.Hat),
	}
	var s []bool
	switch track {
	case "kick":
		s = out.Tracks.Kick
	case "snare":
		s = out.Tracks.Snare
	case "hat":
		s = out.Tracks.Hat
	}
	for i := range s {
		s[i] = !s[i]
	}
	return out, nil
}
