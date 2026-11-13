package synth

import (
	"errors"
	"reflect"
	"testing"
)

func distSteps(on ...int) []bool {
	s := make([]bool, Steps)
	for _, o := range on {
		s[o-1] = true
	}
	return s
}

func distAll() []bool {
	s := make([]bool, Steps)
	for i := range s {
		s[i] = true
	}
	return s
}

func distPattern(kick, snare, hat []bool) Pattern {
	return Pattern{Name: "t", Author: "tester", BPM: 120, Tracks: Tracks{Kick: kick, Snare: snare, Hat: hat}}
}

func distReadme() Pattern {
	return distPattern(distSteps(1, 5, 9, 13), distSteps(5, 13), distSteps(1, 3, 5, 7, 9, 11, 13, 15))
}

func distEmpty() Pattern { return distPattern(distSteps(), distSteps(), distSteps()) }

func distFull() Pattern { return distPattern(distAll(), distAll(), distAll()) }

func distOther() Pattern {
	return distPattern(distSteps(1, 2), distSteps(5), distSteps(2, 4, 6, 8, 10, 12, 14, 16))
}

func TestDistanceValues(t *testing.T) {
	cases := []struct {
		name string
		a, b Pattern
		want int
	}{
		{"self", distReadme(), distReadme(), 0},
		{"empty", distReadme(), distEmpty(), 14},
		{"example", distReadme(), distOther(), 21},
		{"max", distEmpty(), distFull(), 48},
	}
	for _, c := range cases {
		for _, swap := range []bool{false, true} {
			a, b := c.a, c.b
			if swap {
				a, b = b, a
			}
			got, err := Distance(a, b)
			if err != nil || got != c.want {
				t.Errorf("%s swap=%v: got %d, %v; want %d", c.name, swap, got, err, c.want)
			}
		}
	}
}

func TestDistanceIgnoresMeta(t *testing.T) {
	a, b := distReadme(), distReadme()
	b.Name, b.Author, b.BPM, b.Swing = "x", "y", 90, 30
	if got, err := Distance(a, b); err != nil || got != 0 {
		t.Fatalf("got %d, %v", got, err)
	}
}

func TestDistanceNoMutation(t *testing.T) {
	a, b := distReadme(), distOther()
	ac, bc := a, b
	ac.Tracks = Tracks{Kick: append([]bool(nil), a.Tracks.Kick...), Snare: append([]bool(nil), a.Tracks.Snare...), Hat: append([]bool(nil), a.Tracks.Hat...)}
	bc.Tracks = Tracks{Kick: append([]bool(nil), b.Tracks.Kick...), Snare: append([]bool(nil), b.Tracks.Snare...), Hat: append([]bool(nil), b.Tracks.Hat...)}
	if _, err := Distance(a, b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, ac) || !reflect.DeepEqual(b, bc) {
		t.Fatal("Distance mutated its arguments")
	}
}

func TestDistanceInvalid(t *testing.T) {
	badBPM := distReadme()
	badBPM.BPM = 201
	short := distReadme()
	short.Tracks.Kick = short.Tracks.Kick[:15]
	noAuthor := distReadme()
	noAuthor.Author = ""
	cases := []struct {
		name string
		a, b Pattern
	}{
		{"bad b bpm", distReadme(), badBPM},
		{"bad a short", short, distReadme()},
		{"bad a author", noAuthor, distReadme()},
		{"both", short, badBPM},
	}
	for _, c := range cases {
		got, err := Distance(c.a, c.b)
		if got != -1 || !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: got %d, %v", c.name, got, err)
		}
	}
}

func TestDistanceKeepValidate(t *testing.T) {
	p := distReadme()
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	p.BPM = 201
	if err := p.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
