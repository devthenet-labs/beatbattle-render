package synth

// Reverse returns a copy of p with every track played backwards: a track
// plays on step i of the result exactly when it plays on step Steps+1-i of p.
// p is not modified and the result shares no track memory with it. A pattern
// that Validate refuses yields the zero Pattern and that error.
func Reverse(p Pattern) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	out := p
	out.Tracks = Tracks{
		Kick:  reversed(p.Tracks.Kick),
		Snare: reversed(p.Tracks.Snare),
		Hat:   reversed(p.Tracks.Hat),
	}
	return out, nil
}

func reversed(in []bool) []bool {
	out := make([]bool, len(in))
	for i, v := range in {
		out[len(in)-1-i] = v
	}
	return out
}
