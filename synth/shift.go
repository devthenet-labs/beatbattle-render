package synth

// Shift returns a copy of p with every track rotated n steps later,
// wrapping around; a negative n rotates earlier. p is not modified.
func Shift(p Pattern, n int) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	k := n % Steps
	if k < 0 {
		k += Steps
	}
	out := p
	out.Tracks = Tracks{
		Kick:  rotate(p.Tracks.Kick, k),
		Snare: rotate(p.Tracks.Snare, k),
		Hat:   rotate(p.Tracks.Hat, Steps-k),
	}
	return out, nil
}

// rotate returns a fresh slice holding s with s[j] moved to index
// (j+k) mod len(s); s itself is left untouched.
func rotate(s []bool, k int) []bool {
	r := make([]bool, len(s))
	for j, v := range s {
		r[(j+k)%len(s)] = v
	}
	return r
}
