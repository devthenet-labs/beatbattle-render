package synth

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func htSteps(on ...int) []bool {
	s := make([]bool, Steps)
	for _, i := range on {
		s[i] = true
	}
	return s
}

func htPattern(bpm, swing int, kick, snare, hat []bool) Pattern {
	return Pattern{Name: "t", Author: "tester", BPM: bpm, Swing: swing, Tracks: Tracks{Kick: kick, Snare: snare, Hat: hat}}
}

func TestHitTimesValues(t *testing.T) {
	readme := htPattern(120, 0, htSteps(0, 4, 8, 12), htSteps(4, 12), htSteps(0, 2, 4, 6, 8, 10, 12, 14))
	swung := htPattern(100, 30, htSteps(0, 1), htSteps(7), htSteps(14, 15))
	cases := []struct {
		name  string
		p     Pattern
		track string
		want  []int64
	}{
		{"readme kick", readme, "kick", []int64{0, 500, 1000, 1500}},
		{"readme snare", readme, "snare", []int64{500, 1500}},
		{"readme hat", readme, "hat", []int64{0, 250, 500, 750, 1000, 1250, 1500, 1750}},
		{"swing kick", swung, "kick", []int64{0, 195}},
		{"swing snare", swung, "snare", []int64{1095}},
		{"swing hat", swung, "hat", []int64{2100, 2295}},
	}
	for _, c := range cases {
		got, err := HitTimes(c.p, c.track)
		if err != nil || !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, %v; want %v", c.name, got, err, c.want)
		}
	}
}

func TestHitTimesEmpty(t *testing.T) {
	got, err := HitTimes(htPattern(120, 0, htSteps(), htSteps(), htSteps()), "kick")
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("got %#v, %v", got, err)
	}
}

func TestHitTimesErrors(t *testing.T) {
	good := htPattern(120, 0, htSteps(0), htSteps(), htSteps())
	for _, track := range []string{"Kick", "", "tom"} {
		got, err := HitTimes(good, track)
		if got != nil || !errors.Is(err, ErrInvalid) {
			t.Errorf("track %q: got %v, %v", track, got, err)
		}
	}
	bad := htPattern(120, 61, htSteps(0), htSteps(), htSteps())
	got, err := HitTimes(bad, "kick")
	if got != nil || !errors.Is(err, ErrInvalid) {
		t.Errorf("swing 61: got %v, %v", got, err)
	}
}

func TestHitTimesMatchesAnalyze(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	randRow := func() []bool {
		r := make([]bool, Steps)
		for i := range r {
			r[i] = rng.Intn(2) == 0
		}
		return r
	}
	for n := 0; n < 50; n++ {
		p := htPattern(60+rng.Intn(141), rng.Intn(61), randRow(), randRow(), randRow())
		a, err := Analyze(p)
		if err != nil {
			t.Fatal(err)
		}
		rows := map[string][]bool{"kick": p.Tracks.Kick, "snare": p.Tracks.Snare, "hat": p.Tracks.Hat}
		for track, row := range rows {
			got, err := HitTimes(p, track)
			if err != nil {
				t.Fatal(err)
			}
			var want []int64
			for i, on := range row {
				if on {
					want = append(want, a.StepsMs[i])
				}
			}
			if len(want) == 0 {
				want = []int64{}
			}
			if !slices.Equal(got, want) || !slices.IsSorted(got) {
				t.Errorf("%s: got %v want %v", track, got, want)
			}
		}
	}
}
