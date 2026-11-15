package synth

import (
	"bytes"
	"slices"
)

// Retime returns a copy of p with the given bpm and swing. p is not modified
// and the copy shares no slices with it. It returns the zero Pattern and an
// error wrapping ErrInvalid if the bpm, the swing or p is invalid.
func Retime(p Pattern, bpm, swing int) (Pattern, error) {
	q := Pattern{
		ID:     bytes.Clone(p.ID),
		Name:   p.Name,
		Author: p.Author,
		BPM:    bpm,
		Swing:  swing,
		Tracks: Tracks{
			Kick:  slices.Clone(p.Tracks.Kick),
			Snare: slices.Clone(p.Tracks.Snare),
			Hat:   slices.Clone(p.Tracks.Hat),
		},
		Votes:     bytes.Clone(p.Votes),
		CreatedAt: bytes.Clone(p.CreatedAt),
	}
	if err := q.Validate(); err != nil {
		return Pattern{}, err
	}
	return q, nil
}
