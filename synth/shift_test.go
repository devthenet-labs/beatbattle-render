package synth

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

// shiftSteps1 builds a track from 1-based step numbers.
func shiftSteps1(on ...int) []bool {
	s := make([]bool, Steps)
	for _, i := range on {
		s[i-1] = true
	}
	return s
}

func shiftReadme() Pattern {
	return Pattern{
		Name: "readme", Author: "tester", BPM: 120, Swing: 20,
		Tracks: Tracks{
			Kick:  shiftSteps1(1, 5, 9, 13),
			Snare: shiftSteps1(5, 13),
			Hat:   shiftSteps1(1, 3, 5, 7, 9, 11, 13, 15),
		},
	}
}

func shiftMust(t *testing.T, p Pattern, n int) Pattern {
	t.Helper()
	out, err := Shift(p, n)
	if err != nil {
		t.Fatalf("Shift(%d): %v", n, err)
	}
	return out
}

func shiftEqualTracks(t *testing.T, got, want Tracks) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tracks = %v, want %v", got, want)
	}
}

func TestShiftReadme(t *testing.T) {
	evens := shiftSteps1(2, 4, 6, 8, 10, 12, 14, 16)
	cases := []struct {
		n    int
		want Tracks
	}{
		{1, Tracks{shiftSteps1(2, 6, 10, 14), shiftSteps1(6, 14), evens}},
		{3, Tracks{shiftSteps1(4, 8, 12, 16), shiftSteps1(8, 16), shiftSteps1(2, 4, 6, 8, 10, 12, 14, 16)}},
		{-1, Tracks{shiftSteps1(4, 8, 12, 16), shiftSteps1(4, 12), evens}},
	}
	for _, c := range cases {
		shiftEqualTracks(t, shiftMust(t, shiftReadme(), c.n).Tracks, c.want)
	}
}

func TestShiftWrap(t *testing.T) {
	p := shiftReadme()
	p.Tracks.Kick = shiftSteps1(16)
	shiftEqualTracks(t, shiftMust(t, p, 1).Tracks, Tracks{shiftSteps1(1), shiftSteps1(6, 14), shiftSteps1(2, 4, 6, 8, 10, 12, 14, 16)})
}

func TestShiftMultiplesOf16(t *testing.T) {
	p := shiftReadme()
	for _, n := range []int{0, 16, -32, 48} {
		shiftEqualTracks(t, shiftMust(t, p, n).Tracks, p.Tracks)
	}
}

func shiftNaive(s []bool, n int) []bool {
	r := make([]bool, len(s))
	for i := range r {
		r[i] = s[(((i-n)%Steps)+Steps)%Steps]
	}
	return r
}

func TestShiftProperty(t *testing.T) {
	p := shiftReadme()
	ns := []int{-17, -1, 1, 5, 33, 1000003, -1000003, math.MaxInt, math.MinInt}
	for _, a := range ns {
		got := shiftMust(t, p, a).Tracks
		shiftEqualTracks(t, got, Tracks{shiftNaive(p.Tracks.Kick, a), shiftNaive(p.Tracks.Snare, a), shiftNaive(p.Tracks.Hat, a)})
		for _, b := range []int{-3, 7, 21} {
			twice := shiftMust(t, shiftMust(t, p, a), b).Tracks
			shiftEqualTracks(t, twice, Tracks{shiftNaive(p.Tracks.Kick, a+b), shiftNaive(p.Tracks.Snare, a+b), shiftNaive(p.Tracks.Hat, a+b)})
		}
	}
}

func TestShiftKeepsMeta(t *testing.T) {
	p := shiftReadme()
	out := shiftMust(t, p, 5)
	if out.Name != p.Name || out.Author != p.Author || out.BPM != p.BPM || out.Swing != p.Swing {
		t.Fatalf("metadata changed: %+v", out)
	}
}

func TestShiftDoesNotMutate(t *testing.T) {
	p := shiftReadme()
	before := Tracks{append([]bool(nil), p.Tracks.Kick...), append([]bool(nil), p.Tracks.Snare...), append([]bool(nil), p.Tracks.Hat...)}
	out := shiftMust(t, p, 1)
	shiftEqualTracks(t, p.Tracks, before)
	out.Tracks.Kick[0] = !out.Tracks.Kick[0]
	out.Tracks.Snare[0] = !out.Tracks.Snare[0]
	out.Tracks.Hat[0] = !out.Tracks.Hat[0]
	shiftEqualTracks(t, p.Tracks, before)
}

func TestShiftInvalid(t *testing.T) {
	short := shiftReadme()
	short.Tracks.Kick = short.Tracks.Kick[:15]
	badBPM := shiftReadme()
	badBPM.BPM = 10
	for _, p := range []Pattern{short, badBPM} {
		out, err := Shift(p, 1)
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if !reflect.DeepEqual(out, Pattern{}) {
			t.Fatalf("result = %+v, want zero Pattern", out)
		}
	}
}
