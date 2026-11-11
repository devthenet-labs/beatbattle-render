package synth

// Layered returns, in ascending order, the 1-based steps on which at least two
// tracks play. It returns a nil slice and Validate's error (wrapping ErrInvalid)
// when p is invalid, and an empty non-nil slice when no step is layered.
func Layered(p Pattern) ([]int, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	// seen marks the steps an earlier track already plays; a later track
	// hitting one of them makes that step layered.
	var layered [Steps]bool
	seen := p.Tracks.Kick
	for _, track := range [][]bool{p.Tracks.Snare, p.Tracks.Hat} {
		for i, on := range track {
			layered[i] = layered[i] || on && seen[i]
			seen[i] = seen[i] || on
		}
	}
	out := []int{}
	for i, ok := range layered {
		if ok {
			out = append(out, i+1)
		}
	}
	return out, nil
}
