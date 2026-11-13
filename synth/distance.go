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
	n := 0
	for _, t := range [3][2][]bool{
		{a.Tracks.Kick, b.Tracks.Kick},
		{a.Tracks.Snare, b.Tracks.Snare},
		{a.Tracks.Hat, b.Tracks.Hat},
	} {
		for i := 0; i < Steps; i++ {
			if t[0][i] != t[1][i] {
				n++
			}
		}
	}
	return n, nil
}
