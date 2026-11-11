package synth

import (
	"errors"
	"reflect"
	"testing"
)

func onSteps(oneBased ...int) []bool {
	s := make([]bool, Steps)
	for _, n := range oneBased {
		s[n-1] = true
	}
	return s
}

func layeredAll() []int {
	var all []int
	for i := 1; i <= Steps; i++ {
		all = append(all, i)
	}
	return all
}

func layeredPattern(k, s, h []bool) Pattern {
	return Pattern{Name: "t", Author: "tester", BPM: 120, Tracks: Tracks{Kick: k, Snare: s, Hat: h}}
}

func TestLayered(t *testing.T) {
	tests := []struct {
		name string
		p    Pattern
		want []int
	}{
		{"readme", layeredPattern(onSteps(1, 5, 9, 13), onSteps(5, 13), onSteps(1, 3, 5, 7, 9, 11, 13, 15)), []int{1, 5, 9, 13}},
		{"single excluded", layeredPattern(onSteps(2, 7), onSteps(7, 11), onSteps(11, 16)), []int{7, 11}},
		{"all", layeredPattern(onSteps(layeredAll()...), onSteps(layeredAll()...), onSteps(layeredAll()...)), layeredAll()},
		{"empty", layeredPattern(onSteps(), onSteps(), onSteps()), []int{}},
		{"two tracks", layeredPattern(onSteps(3), onSteps(), onSteps(3)), []int{3}},
		{"single track", layeredPattern(onSteps(layeredAll()...), onSteps(), onSteps()), []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Layered(tc.p)
			if err != nil {
				t.Fatal(err)
			}
			if got == nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestLayeredInvalid(t *testing.T) {
	p := layeredPattern(onSteps(1), onSteps(1), onSteps(1))
	p.Author = ""
	short := layeredPattern(onSteps(1), onSteps(1), onSteps(1))
	short.Tracks.Hat = short.Tracks.Hat[:15]
	for name, bad := range map[string]Pattern{"author": p, "short": short} {
		got, err := Layered(bad)
		if got != nil || !errors.Is(err, ErrInvalid) {
			t.Fatalf("%s: got %#v, %v", name, got, err)
		}
	}
}
