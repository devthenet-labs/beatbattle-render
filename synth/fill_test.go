package synth

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"
)

func fillExample() Pattern {
	return Pattern{
		ID: json.RawMessage(`"abc"`), Name: "Example", Author: "tester", BPM: 120, Swing: 10,
		Tracks: Tracks{
			Kick:  steps(0, 4, 8, 12),
			Snare: steps(4, 12),
			Hat:   steps(0, 2, 4, 6, 8, 10, 12, 14),
		},
		Votes:     json.RawMessage(`3`),
		CreatedAt: json.RawMessage(`"2026-01-01T00:00:00Z"`),
	}
}

func TestFillExample(t *testing.T) {
	p := fillExample()
	got, err := Fill(p, "snare", 9, 12)
	if err != nil {
		t.Fatal(err)
	}
	if want := steps(4, 8, 9, 10, 11, 12); !slices.Equal(got.Tracks.Snare, want) {
		t.Errorf("snare = %v, want %v", got.Tracks.Snare, want)
	}
	if !slices.Equal(got.Tracks.Kick, p.Tracks.Kick) || !slices.Equal(got.Tracks.Hat, p.Tracks.Hat) {
		t.Error("kick or hat changed")
	}
	if got.Name != p.Name || got.Author != p.Author || got.BPM != p.BPM || got.Swing != p.Swing {
		t.Error("scalar fields changed")
	}
	if string(got.ID) != string(p.ID) || string(got.Votes) != string(p.Votes) || string(got.CreatedAt) != string(p.CreatedAt) {
		t.Error("raw fields not carried over")
	}
}

func TestFillSingleStep(t *testing.T) {
	p := pattern(120, steps(), steps(), steps())
	got, err := Fill(p, "kick", 16, 16)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Tracks.Kick, steps(15)) {
		t.Errorf("kick = %v", got.Tracks.Kick)
	}
}

func TestFillWholeSpan(t *testing.T) {
	all := make([]int, Steps)
	for i := range all {
		all[i] = i
	}
	for _, track := range []string{"kick", "snare", "hat"} {
		p := fillExample()
		got, err := Fill(p, track, 1, 16)
		if err != nil {
			t.Fatal(err)
		}
		tr := map[string][]bool{"kick": got.Tracks.Kick, "snare": got.Tracks.Snare, "hat": got.Tracks.Hat}
		orig := map[string][]bool{"kick": p.Tracks.Kick, "snare": p.Tracks.Snare, "hat": p.Tracks.Hat}
		for name, s := range tr {
			if name == track {
				if !slices.Equal(s, steps(all...)) {
					t.Errorf("%s not full: %v", name, s)
				}
			} else if !slices.Equal(s, orig[name]) {
				t.Errorf("%s changed", name)
			}
		}
	}
}

func TestFillKeepsOutsideSteps(t *testing.T) {
	p := pattern(120, steps(0, 15), steps(), steps())
	got, err := Fill(p, "kick", 5, 6)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Tracks.Kick, steps(0, 4, 5, 15)) {
		t.Errorf("kick = %v", got.Tracks.Kick)
	}
}

func TestFillNoMutationNoAliasing(t *testing.T) {
	p := fillExample()
	k, s, h := slices.Clone(p.Tracks.Kick), slices.Clone(p.Tracks.Snare), slices.Clone(p.Tracks.Hat)
	got, err := Fill(p, "snare", 1, 16)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got.Tracks.Kick {
		got.Tracks.Kick[i] = !got.Tracks.Kick[i]
		got.Tracks.Snare[i] = !got.Tracks.Snare[i]
		got.Tracks.Hat[i] = !got.Tracks.Hat[i]
	}
	got.ID[0] = 'X'
	got.Votes[0] = 'X'
	got.CreatedAt[0] = 'X'
	if !slices.Equal(p.Tracks.Kick, k) || !slices.Equal(p.Tracks.Snare, s) || !slices.Equal(p.Tracks.Hat, h) {
		t.Error("input tracks modified")
	}
	if string(p.ID) != `"abc"` || string(p.Votes) != `3` || string(p.CreatedAt) != `"2026-01-01T00:00:00Z"` {
		t.Error("input raw fields modified")
	}
}

func TestFillErrors(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(*Pattern)
		track    string
		from, to int
	}{
		{"from 0", nil, "kick", 0, 4},
		{"to 17", nil, "kick", 1, 17},
		{"from 17", nil, "kick", 17, 17},
		{"to 0", nil, "kick", 1, 0},
		{"from > to", nil, "kick", 5, 4},
		{"track Hat", nil, "Hat", 1, 4},
		{"track empty", nil, "", 1, 4},
		{"track snare space", nil, "snare ", 1, 4},
		{"track drums", nil, "drums", 1, 4},
		{"empty author", func(p *Pattern) { p.Author = "" }, "kick", 1, 4},
		{"bpm 0", func(p *Pattern) { p.BPM = 0 }, "kick", 1, 4},
		{"short track", func(p *Pattern) { p.Tracks.Hat = p.Tracks.Hat[:15] }, "kick", 1, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := fillExample()
			if c.mutate != nil {
				c.mutate(&p)
			}
			got, err := Fill(p, c.track, c.from, c.to)
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
			if !reflect.DeepEqual(got, Pattern{}) {
				t.Errorf("got %+v, want zero Pattern", got)
			}
		})
	}
}
