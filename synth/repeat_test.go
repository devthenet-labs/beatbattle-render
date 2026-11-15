package synth

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"slices"
	"testing"
)

func trackOf(p Pattern, name string) []bool {
	switch name {
	case "kick":
		return p.Tracks.Kick
	case "snare":
		return p.Tracks.Snare
	}
	return p.Tracks.Hat
}

func readme() Pattern {
	return pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6, 8, 10, 12, 14))
}

func onSteps(s []bool) []int {
	var r []int
	for i, b := range s {
		if b {
			r = append(r, i)
		}
	}
	return r
}

func TestRepeatExamples(t *testing.T) {
	cases := []struct {
		name   string
		p      Pattern
		track  string
		length int
		want   []int
	}{
		{"snare 6", readme(), "snare", 6, []int{4, 10}},
		{"kick step 2 len 3", pattern(120, steps(1), steps(), steps()), "kick", 3, []int{1, 4, 7, 10, 13}},
		{"len 16 unchanged", readme(), "hat", 16, []int{0, 2, 4, 6, 8, 10, 12, 14}},
		{"len 1 on", readme(), "kick", 1, onSteps(steps(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))},
		{"len 1 off", pattern(120, steps(1), steps(), steps()), "kick", 1, nil},
	}
	for _, c := range cases {
		out, err := Repeat(c.p, c.track, c.length)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := onSteps(trackOf(out, c.track)); !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
	out, _ := Repeat(readme(), "snare", 6)
	if !slices.Equal(out.Tracks.Kick, readme().Tracks.Kick) || !slices.Equal(out.Tracks.Hat, readme().Tracks.Hat) {
		t.Error("other tracks changed")
	}
}

func TestRepeatProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for n := 0; n < 5; n++ {
		mk := func() []bool {
			s := make([]bool, Steps)
			for i := range s {
				s[i] = rng.Intn(2) == 0
			}
			return s
		}
		p := pattern(100, mk(), mk(), mk())
		for _, tr := range []string{"kick", "snare", "hat"} {
			for l := 1; l <= Steps; l++ {
				out, err := Repeat(p, tr, l)
				if err != nil {
					t.Fatal(err)
				}
				in, got := trackOf(p, tr), trackOf(out, tr)
				if len(got) != Steps {
					t.Fatalf("len %d", len(got))
				}
				for i := range got {
					if got[i] != in[i%l] {
						t.Fatalf("%s %d step %d", tr, l, i)
					}
				}
				for _, o := range []string{"kick", "snare", "hat"} {
					if o != tr && !slices.Equal(trackOf(out, o), trackOf(p, o)) {
						t.Fatalf("%s changed", o)
					}
				}
			}
		}
	}
}

func TestRepeatErrors(t *testing.T) {
	badSwing := readme()
	badSwing.Swing = 61
	short := readme()
	short.Tracks.Kick = short.Tracks.Kick[:15]
	cases := []struct {
		name   string
		p      Pattern
		track  string
		length int
	}{
		{"len 0", readme(), "kick", 0},
		{"len -1", readme(), "kick", -1},
		{"len 17", readme(), "kick", 17},
		{"Snare", readme(), "Snare", 4},
		{"empty", readme(), "", 4},
		{"kicks", readme(), "kicks", 4},
		{"swing 61", badSwing, "kick", 4},
		{"15 steps", short, "snare", 4},
	}
	for _, c := range cases {
		out, err := Repeat(c.p, c.track, c.length)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err %v", c.name, err)
		}
		if !reflect.DeepEqual(out, Pattern{}) {
			t.Errorf("%s: not zero pattern", c.name)
		}
	}
}

func TestRepeatNoMutationAndFields(t *testing.T) {
	p := readme()
	p.ID = json.RawMessage(`"abc"`)
	p.Votes = json.RawMessage(`3`)
	p.CreatedAt = json.RawMessage(`"2026-01-01"`)
	p.Swing = 20
	snap := Tracks{Kick: slices.Clone(p.Tracks.Kick), Snare: slices.Clone(p.Tracks.Snare), Hat: slices.Clone(p.Tracks.Hat)}
	out, err := Repeat(p, "snare", 6)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Tracks, snap) {
		t.Error("p modified")
	}
	if out.Name != p.Name || out.Author != p.Author || out.BPM != p.BPM || out.Swing != p.Swing ||
		string(out.ID) != string(p.ID) || string(out.Votes) != string(p.Votes) || string(out.CreatedAt) != string(p.CreatedAt) {
		t.Error("fields differ")
	}
	for _, tr := range [][]bool{out.Tracks.Kick, out.Tracks.Snare, out.Tracks.Hat} {
		if len(tr) != Steps {
			t.Errorf("len %d", len(tr))
		}
		for i := range tr {
			tr[i] = !tr[i]
		}
	}
	if !reflect.DeepEqual(p.Tracks, snap) {
		t.Error("mutating result changed p")
	}
}
