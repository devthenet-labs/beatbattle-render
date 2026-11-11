package synth

import "testing"

func restRow(on ...int) []bool {
	s := make([]bool, Steps)
	for _, i := range on {
		s[i-1] = true
	}
	return s
}

func restPattern(bpm int, kick, snare, hat []bool) Pattern {
	return Pattern{Name: "t", Author: "tester", BPM: bpm, Tracks: Tracks{Kick: kick, Snare: snare, Hat: hat}}
}

func TestLongestRest(t *testing.T) {
	readme := restPattern(120, restRow(1, 5, 9, 13), restRow(5, 13), restRow(1, 3, 5, 7, 9, 11, 13, 15))
	all := restRow(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
	cases := []struct {
		name  string
		p     Pattern
		track string
		want  int
	}{
		{"readme kick", readme, "kick", 3},
		{"readme snare", readme, "snare", 7},
		{"readme hat", readme, "hat", 1},
		{"empty", restPattern(120, restRow(), restRow(), restRow()), "kick", 16},
		{"full", restPattern(120, all, restRow(), restRow()), "kick", 0},
		{"only first", restPattern(120, restRow(1), restRow(), restRow()), "kick", 15},
		{"only last", restPattern(120, restRow(16), restRow(), restRow()), "kick", 15},
		{"first and last", restPattern(120, restRow(1, 16), restRow(), restRow()), "kick", 14},
		{"other tracks ignored", restPattern(120, restRow(1, 16), all, all), "kick", 14},
		{"clap", readme, "clap", -1},
		{"Kick", readme, "Kick", -1},
		{"empty name", readme, "", -1},
		{"space kick", readme, " kick", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := LongestRest(tc.p, tc.track); got != tc.want {
				t.Fatalf("LongestRest(%q) = %d, want %d", tc.track, got, tc.want)
			}
		})
	}
}

func TestLongestRestInvalidPattern(t *testing.T) {
	short := restPattern(120, restRow(1)[:15], restRow(), restRow())
	for name, p := range map[string]Pattern{
		"bpm 59":        restPattern(59, restRow(1), restRow(), restRow()),
		"15 step track": short,
	} {
		for _, track := range []string{"kick", "snare", "hat"} {
			if got := LongestRest(p, track); got != -1 {
				t.Errorf("%s: LongestRest(%q) = %d, want -1", name, track, got)
			}
		}
	}
}
