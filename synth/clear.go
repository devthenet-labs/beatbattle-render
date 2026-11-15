package synth

import "slices"

// Clear returns a deep copy of p in which every step of the named track
// ("kick", "snare" or "hat") is a rest. p is not modified and the copy shares
// no memory with it. It returns the zero Pattern and an error wrapping
// ErrInvalid if p fails Validate or the track name is unknown.
func Clear(p Pattern, track string) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
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
		out.Tracks.Kick = make([]bool, Steps)
	case "snare":
		out.Tracks.Snare = make([]bool, Steps)
	case "hat":
		out.Tracks.Hat = make([]bool, Steps)
	default:
		return Pattern{}, invalid("unknown track %q", track)
	}
	return out, nil
}
