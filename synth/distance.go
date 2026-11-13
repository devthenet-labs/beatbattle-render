package synth

// Distance counts the (track, step) pairs where exactly one of a and b plays,
// a value from 0 to 48. Name, author, bpm and swing are ignored. It returns -1
// and an error wrapping ErrInvalid if Validate refuses either pattern.
func Distance(a, b Pattern) (int, error) {
	if err := a.Validate(); err != nil {
		return -1, err
	}
	if sharesTracks(a.Tracks, b.Tracks) {
		// b reads the very same steps as a (for example a copy of a with
		// other metadata), so nothing can differ.
		return 0, nil
	}
	if err := b.Validate(); err != nil {
		return -1, err
	}
	return countDiff(a.Tracks.Kick, b.Tracks.Kick) +
		countDiff(a.Tracks.Snare, b.Tracks.Snare) +
		countDiff(a.Tracks.Hat, b.Tracks.Hat), nil
}

// countDiff counts the steps on which exactly one of x and y plays.
func countDiff(x, y []bool) int {
	n := 0
	for i := 0; i < Steps; i++ {
		if x[i] != y[i] {
			n++
		}
	}
	return n
}

// sharesTracks reports whether x and y are backed by the same step arrays.
func sharesTracks(x, y Tracks) bool {
	same := func(p, q []bool) bool {
		return len(p) == len(q) && len(p) > 0 && &p[0] == &q[0]
	}
	return same(x.Kick, y.Kick) && same(x.Snare, y.Snare) && same(x.Hat, y.Hat)
}
