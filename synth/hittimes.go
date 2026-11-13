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
	row, err := trackRow(p.Tracks, track)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, Steps)
	for i, on := range row {
		if on {
			out = append(out, hitMs(i, p.BPM, p.Swing))
		}
	}
	return out, nil
}

// trackRow selects a track's steps by its exact name.
func trackRow(t Tracks, track string) ([]bool, error) {
	switch track {
	case "kick":
		return t.Kick, nil
	case "snare":
		return t.Snare, nil
	case "hat":
		return t.Hat, nil
	}
	return nil, invalid("unknown track %q", track)
}

// hitMs is the start of step i in ms: the straight sixteenth plus, on the
// off-beats (steps 2, 4, 6, ... counting from 1), the swing delay.
func hitMs(i, bpm, swing int) int64 {
	ms := DurationMs(stepStart(i, bpm, 0))
	if i%2 == 0 {
		ms += DurationMs((swing*SampleRate*15 + 50*bpm) / (100 * bpm))
	}
	return ms
}
