package synth

import "slices"

// Repeat returns a copy of p in which the named track ("kick", "snare" or
// "hat") repeats its first length steps over all Steps steps: for every step i
// from 1 to 16, the copy's track plays on step i exactly when p's track plays
// on step ((i-1) mod length)+1. The other tracks and fields are as in p, and p
// is not modified. A length outside 1 to 16, an unknown track name or a pattern
// that Validate refuses yields the zero Pattern and an error wrapping
// ErrInvalid.
func Repeat(p Pattern, track string, length int) (Pattern, error) {
	if length < 1 || length > Steps {
		return Pattern{}, invalid("length must be an integer from 1 to %d", Steps)
	}
	switch track {
	case "kick", "snare", "hat":
	default:
		return Pattern{}, invalid("track must be kick, snare or hat")
	}
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	out := clonePattern(p)
	if length == Steps {
		// Repeating all the steps leaves the track as it is.
		return out, nil
	}
	dst := out.Tracks.Hat
	switch track {
	case "kick":
		dst = out.Tracks.Kick
	case "snare":
		dst = out.Tracks.Snare
	}
	for i := range dst {
		dst[i] = dst[i%length]
	}
	return out, nil
}

// clonePattern returns a copy of p whose track slices share no backing array
// with p's.
func clonePattern(p Pattern) Pattern {
	out := p
	out.Tracks = Tracks{
		Kick:  slices.Clone(p.Tracks.Kick),
		Snare: slices.Clone(p.Tracks.Snare),
		Hat:   slices.Clone(p.Tracks.Hat),
	}
	return out
}
