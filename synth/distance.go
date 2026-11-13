package synth

// Distance counts the (track, step) pairs where exactly one of a and b plays,
// a value from 0 to 48. Name, author, bpm and swing are ignored. It returns -1
// and an error wrapping ErrInvalid if Validate refuses either pattern.
func Distance(a, b Pattern) (int, error) {
	if err := a.Validate(); err != nil {
		return -1, err
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
