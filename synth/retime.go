package synth

import (
	"encoding/json"
	"slices"
)

// Retime returns a copy of p with the given bpm and swing. p is not modified
// and the copy shares no slices with it. It returns the zero Pattern and an
// error wrapping ErrInvalid if the bpm, the swing or p is invalid.
func Retime(p Pattern, bpm, swing int) (Pattern, error) {
	q := Pattern{
		ID:     p.ID,
		Name:   p.Name,
		Author: p.Author,
		BPM:    bpm,
		Swing:  swing,
		Tracks: Tracks{
			Kick:  slices.Clone(p.Tracks.Kick),
			Snare: slices.Clone(p.Tracks.Snare),
			Hat:   slices.Clone(p.Tracks.Hat),
		},
		Votes:     p.Votes,
		CreatedAt: p.CreatedAt,
	}
	cloneRaw(&q.ID, &q.Votes, &q.CreatedAt)
	if err := q.Validate(); err != nil {
		return Pattern{}, err
	}
	return q, nil
}

// cloneRaw replaces each stored field with a copy, using one allocation for
// all of them. A pattern with no stored fields is left as it is.
func cloneRaw(fields ...*json.RawMessage) {
	n, set := 0, false
	for _, f := range fields {
		n += len(*f)
		set = set || *f != nil
	}
	if !set {
		return
	}
	buf := make([]byte, 0, n)
	for _, f := range fields {
		start := len(buf)
		buf = append(buf, *f...)
		*f = buf[start:len(buf):len(buf)]
	}
}
