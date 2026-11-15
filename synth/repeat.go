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
	var src []bool
	switch track {
	case "kick":
		src = p.Tracks.Kick
	case "snare":
		src = p.Tracks.Snare
	default:
		src = p.Tracks.Hat
	}
	rep := make([]bool, Steps)
	for i := range rep {
		rep[i] = src[i%length]
	}
	out := p
	out.Tracks = Tracks{
		Kick:  slices.Clone(p.Tracks.Kick),
		Snare: slices.Clone(p.Tracks.Snare),
		Hat:   slices.Clone(p.Tracks.Hat),
	}
	switch track {
	case "kick":
		out.Tracks.Kick = rep
	case "snare":
		out.Tracks.Snare = rep
	default:
		out.Tracks.Hat = rep
	}
	return out, nil
}
