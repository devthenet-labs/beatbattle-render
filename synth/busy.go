package synth

// Busy counts how many of a pattern's 3*Steps steps play, or returns -1 if
// Validate refuses the pattern.
func Busy(p Pattern) int {
	if err := p.Validate(); err != nil {
		return -1
	}
	n := 0
	for _, track := range [][]bool{p.Tracks.Kick, p.Tracks.Snare, p.Tracks.Hat} {
		for _, on := range track {
			if on {
				n++
			}
		}
	}
	return n
}
