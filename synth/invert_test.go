package synth

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func readme() Pattern {
	return pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6, 8, 10, 12, 14))
}

func TestInvertReadme(t *testing.T) {
	p := readme()
	got, err := Invert(p, "hat")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Tracks.Hat, steps(1, 3, 5, 7, 9, 11, 13, 15)) {
		t.Errorf("hat = %v", got.Tracks.Hat)
	}
	if !reflect.DeepEqual(got.Tracks.Kick, p.Tracks.Kick) || !reflect.DeepEqual(got.Tracks.Snare, p.Tracks.Snare) {
		t.Error("other tracks changed")
	}
	got, err = Invert(p, "snare")
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range got.Tracks.Snare {
		if v != (i != 4 && i != 12) {
			t.Errorf("snare step %d = %v", i, v)
		}
	}
}

func pick(p *Pattern, track string) []bool {
	switch track {
	case "kick":
		return p.Tracks.Kick
	case "snare":
		return p.Tracks.Snare
	}
	return p.Tracks.Hat
}

func TestInvertTracksAndFields(t *testing.T) {
	p := readme()
	p.Name, p.Author, p.Swing = "n", "a", 30
	p.ID, p.Votes, p.CreatedAt = json.RawMessage(`"x"`), json.RawMessage(`3`), json.RawMessage(`"t"`)
	for _, tr := range []string{"kick", "snare", "hat"} {
		got, err := Invert(p, tr)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range []string{"kick", "snare", "hat"} {
			in, out := pick(&p, o), pick(&got, o)
			for i := range in {
				if (o == tr) == (in[i] == out[i]) {
					t.Fatalf("%s/%s step %d", tr, o, i)
				}
			}
		}
		if got.Name != p.Name || got.Author != p.Author || got.BPM != p.BPM || got.Swing != p.Swing ||
			string(got.ID) != `"x"` || string(got.Votes) != `3` || string(got.CreatedAt) != `"t"` {
			t.Errorf("fields changed: %+v", got)
		}
	}
}

func TestInvertTwiceRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	rnd := func() []bool {
		s := make([]bool, Steps)
		for i := range s {
			s[i] = r.Intn(2) == 0
		}
		return s
	}
	for n := 0; n < 100; n++ {
		p := pattern(100, rnd(), rnd(), rnd())
		for _, tr := range []string{"kick", "snare", "hat"} {
			a, err := Invert(p, tr)
			if err != nil {
				t.Fatal(err)
			}
			b, err := Invert(a, tr)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(b.Tracks, p.Tracks) {
				t.Fatal("double invert differs")
			}
		}
	}
}

func TestInvertNoMutateNoAlias(t *testing.T) {
	p := readme()
	p.ID = json.RawMessage(`"x"`)
	snap := Tracks{Kick: append([]bool(nil), p.Tracks.Kick...), Snare: append([]bool(nil), p.Tracks.Snare...), Hat: append([]bool(nil), p.Tracks.Hat...)}
	got, err := Invert(p, "kick")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Tracks, snap) {
		t.Fatal("input mutated")
	}
	for _, s := range [][]bool{got.Tracks.Kick, got.Tracks.Snare, got.Tracks.Hat} {
		for i := range s {
			s[i] = !s[i]
		}
	}
	got.ID[1] = 'y'
	if !reflect.DeepEqual(p.Tracks, snap) || string(p.ID) != `"x"` {
		t.Fatal("result aliases input")
	}
}

func TestInvertErrors(t *testing.T) {
	for _, tr := range []string{"Kick", "cymbal", "", " kick"} {
		got, err := Invert(readme(), tr)
		if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("track %q: %v %v", tr, got, err)
		}
	}
	bad := readme()
	bad.BPM = 201
	short := readme()
	short.Tracks.Kick = short.Tracks.Kick[:15]
	for _, p := range []Pattern{bad, short} {
		got, err := Invert(p, "kick")
		if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("%v %v", got, err)
		}
	}
}
