package synth

import (
	"errors"
	"strings"
	"testing"
)

func gridTracks(on ...int) []bool {
	s := make([]bool, Steps)
	for _, n := range on {
		s[n-1] = true
	}
	return s
}

func gridPattern(k, s, h []bool) Pattern {
	return Pattern{Name: "n", Author: "a", BPM: 120, Tracks: Tracks{Kick: k, Snare: s, Hat: h}}
}

func TestGridREADME(t *testing.T) {
	p := gridPattern(gridTracks(1, 5, 9, 13), gridTracks(5, 13), gridTracks(1, 3, 5, 7, 9, 11, 13, 15))
	got, err := Grid(p)
	if err != nil {
		t.Fatal(err)
	}
	want := "kick  x...x...x...x...\nsnare ....x.......x...\nhat   x.x.x.x.x.x.x.x.\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGridEmptyAndFull(t *testing.T) {
	all := make([]int, Steps)
	for i := range all {
		all[i] = i + 1
	}
	for _, c := range []struct {
		ch string
		on []int
	}{{".", nil}, {"x", all}} {
		got, err := Grid(gridPattern(gridTracks(c.on...), gridTracks(c.on...), gridTracks(c.on...)))
		if err != nil {
			t.Fatal(err)
		}
		want := "kick  " + strings.Repeat(c.ch, 16) + "\nsnare " + strings.Repeat(c.ch, 16) + "\nhat   " + strings.Repeat(c.ch, 16) + "\n"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
		if !strings.HasSuffix(got, "\n") || strings.Count(got, "\n") != 3 {
			t.Fatalf("bad line structure %q", got)
		}
		for _, l := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
			if len(l) != 22 {
				t.Fatalf("line %q has length %d", l, len(l))
			}
		}
	}
}

func TestGridInvalid(t *testing.T) {
	bad := gridPattern(gridTracks(1), make([]bool, 17), gridTracks())
	short := gridPattern(make([]bool, 15), gridTracks(), gridTracks())
	bpm := gridPattern(gridTracks(), gridTracks(), gridTracks())
	bpm.BPM = 10
	for name, p := range map[string]Pattern{"long snare": bad, "short kick": short, "bpm": bpm} {
		got, err := Grid(p)
		if got != "" || !errors.Is(err, ErrInvalid) {
			t.Fatalf("%s: got %q, %v", name, got, err)
		}
	}
}
