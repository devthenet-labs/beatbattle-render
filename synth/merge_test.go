package synth

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

// on1 returns a Steps-long track with the given 1-based steps set.
func on1(on ...int) []bool {
	s := make([]bool, Steps)
	for _, i := range on {
		s[i-1] = true
	}
	return s
}

func mergePat(name, author string, bpm, swing int, kick, snare, hat []bool) Pattern {
	return Pattern{Name: name, Author: author, BPM: bpm, Swing: swing, Tracks: Tracks{Kick: kick, Snare: snare, Hat: hat}}
}

func TestMergeTracks(t *testing.T) {
	cases := []struct {
		name             string
		a, b             Pattern
		kick, snare, hat []int
	}{
		{
			"readme",
			mergePat("a", "alice", 120, 0, on1(1, 5, 9, 13), on1(5, 13), on1(1, 3, 5, 7, 9, 11, 13, 15)),
			mergePat("b", "bob", 120, 0, on1(2, 16), on1(1), on1(4, 10)),
			[]int{1, 2, 5, 9, 13, 16}, []int{1, 5, 13}, []int{1, 3, 4, 5, 7, 9, 10, 11, 13, 15},
		},
		{
			"sparse",
			mergePat("a", "alice", 120, 0, on1(3), on1(7, 8), on1()),
			mergePat("b", "bob", 120, 0, on1(3, 12), on1(), on1(16)),
			[]int{3, 12}, []int{7, 8}, []int{16},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Merge(c.a, c.b)
			if err != nil {
				t.Fatal(err)
			}
			checkSteps(t, "kick", got.Tracks.Kick, c.kick)
			checkSteps(t, "snare", got.Tracks.Snare, c.snare)
			checkSteps(t, "hat", got.Tracks.Hat, c.hat)
		})
	}
}

// checkSteps reports any step of got that plays but is not listed in want
// (1-based).
func checkSteps(t *testing.T, track string, got []bool, want []int) {
	t.Helper()
	if len(got) != Steps {
		t.Fatalf("%s: %d steps, want %d", track, len(got), Steps)
	}
	for i, on := range got {
		if on && !slices.Contains(want, i+1) {
			t.Errorf("%s: step %d plays, want only %v", track, i+1, want)
		}
	}
}

func TestMergeMetadataFromA(t *testing.T) {
	a := mergePat("first", "alice", 120, 0, on1(), on1(), on1())
	b := mergePat("second", "bob", 90, 30, on1(), on1(), on1())
	got, err := Merge(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "first" || got.Author != "alice" || got.BPM != 120 || got.Swing != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestMergeDoesNotMutateOrAlias(t *testing.T) {
	a := mergePat("a", "alice", 120, 0, on1(1, 5), on1(5), on1(1, 3))
	b := mergePat("b", "bob", 100, 10, on1(2), on1(1), on1(4))
	wantA := mergePat("a", "alice", 120, 0, on1(1, 5), on1(5), on1(1, 3))
	wantB := mergePat("b", "bob", 100, 10, on1(2), on1(1), on1(4))
	got, err := Merge(a, b)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range [][]bool{got.Tracks.Kick, got.Tracks.Snare, got.Tracks.Hat} {
		for i := range tr {
			tr[i] = !tr[i]
		}
	}
	if !reflect.DeepEqual(a, wantA) || !reflect.DeepEqual(b, wantB) {
		t.Errorf("inputs modified: a=%+v b=%+v", a, b)
	}
}

func TestMergeInvalid(t *testing.T) {
	good := mergePat("g", "alice", 120, 0, on1(1), on1(2), on1(3))
	badHat := mergePat("b", "bob", 120, 0, on1(1), on1(2), make([]bool, 15))
	badBPM := mergePat("a", "alice", 10, 0, on1(1), on1(2), on1(3))
	for name, args := range map[string][2]Pattern{"invalid b": {good, badHat}, "invalid a": {badBPM, good}} {
		t.Run(name, func(t *testing.T) {
			got, err := Merge(args[0], args[1])
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
			if !reflect.DeepEqual(got, Pattern{}) {
				t.Errorf("got %+v, want zero Pattern", got)
			}
		})
	}
}
