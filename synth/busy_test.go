package synth

import "testing"

func TestBusy(t *testing.T) {
	cases := []struct {
		name string
		p    Pattern
		want int
	}{
		{
			"readme example",
			pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6, 8, 10, 12, 14)),
			14,
		},
		{"no steps", pattern(120, steps(), steps(), steps()), 0},
		{
			"invalid: short hat track",
			func() Pattern {
				p := pattern(120, steps(), steps(), steps())
				p.Tracks.Hat = make([]bool, 15)
				return p
			}(),
			-1,
		},
		{
			"invalid: bpm out of range",
			func() Pattern {
				p := pattern(120, steps(), steps(), steps())
				p.BPM = 0
				return p
			}(),
			-1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Busy(tc.p); got != tc.want {
				t.Errorf("Busy() = %d, want %d", got, tc.want)
			}
		})
	}
}
