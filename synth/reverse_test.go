package synth

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestReverseReadme(t *testing.T) {
	var hat []int
	for i := 0; i < Steps; i += 2 {
		hat = append(hat, i)
	}
	got, err := Reverse(pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(hat...)))
	if err != nil {
		t.Fatal(err)
	}
	var evens []int
	for i := 1; i < Steps; i += 2 {
		evens = append(evens, i)
	}
	if !slices.Equal(got.Tracks.Kick, steps(3, 7, 11, 15)) ||
		!slices.Equal(got.Tracks.Snare, steps(3, 11)) ||
		!slices.Equal(got.Tracks.Hat, steps(evens...)) {
		t.Fatalf("unexpected tracks: %+v", got.Tracks)
	}
}

func TestReverseSecondExample(t *testing.T) {
	got, err := Reverse(pattern(120, steps(0, 1), steps(2), steps(15)))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Tracks.Kick, steps(14, 15)) ||
		!slices.Equal(got.Tracks.Snare, steps(13)) ||
		!slices.Equal(got.Tracks.Hat, steps(0)) {
		t.Fatalf("unexpected tracks: %+v", got.Tracks)
	}
}

func TestReverseTwice(t *testing.T) {
	p := pattern(120, steps(0, 5, 9), steps(2, 3), steps(1, 15))
	once, err := Reverse(p)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := Reverse(once)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(twice.Tracks, p.Tracks) {
		t.Fatalf("double reverse differs: %+v", twice.Tracks)
	}
}

func TestReverseKeepsFields(t *testing.T) {
	p := pattern(97, steps(0), steps(), steps())
	p.Swing = 30
	got, err := Reverse(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != p.Name || got.Author != p.Author || got.BPM != p.BPM || got.Swing != p.Swing {
		t.Fatalf("fields changed: %+v", got)
	}
}

func TestReverseNoMutation(t *testing.T) {
	p := pattern(120, steps(0, 7), steps(3), steps(15))
	kick, snare, hat := slices.Clone(p.Tracks.Kick), slices.Clone(p.Tracks.Snare), slices.Clone(p.Tracks.Hat)
	got, err := Reverse(p)
	if err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		if !slices.Equal(p.Tracks.Kick, kick) || !slices.Equal(p.Tracks.Snare, snare) || !slices.Equal(p.Tracks.Hat, hat) {
			t.Fatal("input was modified")
		}
	}
	check()
	for _, s := range [][]bool{got.Tracks.Kick, got.Tracks.Snare, got.Tracks.Hat} {
		for i := range s {
			s[i] = !s[i]
		}
	}
	check()
}

func TestReverseInvalid(t *testing.T) {
	cases := map[string]func(*Pattern){
		"hat 17 steps": func(p *Pattern) { p.Tracks.Hat = append(p.Tracks.Hat, false) },
		"bpm 59":       func(p *Pattern) { p.BPM = 59 },
		"empty name":   func(p *Pattern) { p.Name = "" },
		"long name":    func(p *Pattern) { p.Name = strings.Repeat("x", 41) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := pattern(120, steps(), steps(), steps())
			mutate(&p)
			got, err := Reverse(p)
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if !reflect.DeepEqual(got, Pattern{}) {
				t.Fatalf("got %+v, want zero Pattern", got)
			}
		})
	}
}
