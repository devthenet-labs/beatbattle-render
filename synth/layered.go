package synth

// Layered returns, in ascending order, the 1-based steps on which at least two
// tracks play. It returns a nil slice and Validate's error (wrapping ErrInvalid)
// when p is invalid, and an empty non-nil slice when no step is layered.
func Layered(p Pattern) ([]int, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	out := []int{}
	for i := 0; i < Steps; i++ {
		n := 0
		for _, on := range []bool{p.Tracks.Kick[i], p.Tracks.Snare[i], p.Tracks.Hat[i]} {
			if on {
				n++
			}
		}
		if n >= 2 {
			out = append(out, i+1)
		}
	}
	return out, nil
}
