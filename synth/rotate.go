package synth

import "slices"

// Rotate returns a copy of p in which the named track ("kick", "snare" or
// "hat") is moved n steps later, wrapping around the end: step i of the copy
// plays exactly when step ((i-1-n) mod 16)+1 of p's track plays. Negative n
// moves the track earlier. p is not modified and the copy shares no memory
// with it. An unknown track name or a pattern that Validate refuses yields the
// zero Pattern and an error wrapping ErrInvalid.
func Rotate(p Pattern, track string, n int) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	var src []bool
	switch track {
	case "kick":
		src = p.Tracks.Kick
	case "snare":
		src = p.Tracks.Snare
	case "hat":
		src = p.Tracks.Hat
	default:
		return Pattern{}, invalid("unknown track %q", track)
	}

	m := n % Steps
	rotated := make([]bool, Steps)
	for i := range rotated {
		rotated[i] = src[((i-m)%Steps+Steps)%Steps]
	}

	out := p
	out.ID = slices.Clone(p.ID)
	out.Votes = slices.Clone(p.Votes)
	out.CreatedAt = slices.Clone(p.CreatedAt)
	out.Tracks = Tracks{
		Kick:  slices.Clone(p.Tracks.Kick),
		Snare: slices.Clone(p.Tracks.Snare),
		Hat:   slices.Clone(p.Tracks.Hat),
	}
	switch track {
	case "kick":
		out.Tracks.Kick = rotated
	case "snare":
		out.Tracks.Snare = rotated
	case "hat":
		out.Tracks.Hat = rotated
	}
	return out, nil
}
