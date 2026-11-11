package synth

import "strings"

// Grid draws p as three text lines, one per track (kick, snare, hat), with
// "x" for a playing step and "." for a silent one. It returns an error
// wrapping ErrInvalid if p fails Validate.
func Grid(p Pattern) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, t := range []struct {
		name  string
		steps []bool
	}{{"kick  ", p.Tracks.Kick}, {"snare ", p.Tracks.Snare}, {"hat   ", p.Tracks.Hat}} {
		b.WriteString(t.name)
		for i := 0; i < Steps; i++ {
			if t.steps[i] {
				b.WriteByte('x')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}
