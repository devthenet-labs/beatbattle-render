package synth

// HitTimes returns, in ascending order, the swung start time in whole
// milliseconds of every step on which the named track ("kick", "snare" or
// "hat") plays. A track with no playing step yields an empty non-nil slice.
// An unknown track name or an invalid pattern yields a nil slice and an error
// wrapping ErrInvalid.
func HitTimes(p Pattern, track string) ([]int64, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	var row []bool
	switch track {
	case "kick":
		row = p.Tracks.Kick
	case "snare":
		row = p.Tracks.Snare
	case "hat":
		row = p.Tracks.Hat
	default:
		return nil, invalid("unknown track %q", track)
	}
	out := make([]int64, 0, Steps)
	for i, on := range row {
		if on {
			out = append(out, DurationMs(stepStart(i, p.BPM, p.Swing)))
		}
	}
	return out, nil
}
