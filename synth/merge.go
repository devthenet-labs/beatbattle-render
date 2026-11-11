package synth

import "slices"

// Merge combines two patterns: a step plays in the result when it plays in a
// or in b. Name, author, BPM and swing come from a. Neither input is modified
// and the result shares no track slices with them. If either pattern fails
// Validate, Merge returns the zero Pattern and an error wrapping ErrInvalid.
func Merge(a, b Pattern) (Pattern, error) {
	if err := a.Validate(); err != nil {
		return Pattern{}, err
	}
	if err := b.Validate(); err != nil {
		return Pattern{}, err
	}
	return Pattern{
		Name: a.Name, Author: a.Author, BPM: a.BPM, Swing: a.Swing,
		Tracks: Tracks{
			Kick:  orSteps(a.Tracks.Kick, b.Tracks.Kick),
			Snare: orSteps(a.Tracks.Snare, b.Tracks.Snare),
			Hat:   orSteps(a.Tracks.Hat, b.Tracks.Hat),
		},
	}, nil
}

// orSteps lays y's hits over a copy of x.
func orSteps(x, y []bool) []bool {
	out := slices.Clone(x)
	for i, hit := range y {
		if hit {
			out[i] = true
		}
	}
	return out
}
