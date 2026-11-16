package synth

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func readme() Pattern {
	return pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6, 8, 10, 12, 14))
}

func TestToggleReadme(t *testing.T) {
	p := readme()
	got, err := Toggle(p, []Edit{{"snare", 5}, {"snare", 8}, {"hat", 2}, {"hat", 2}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Tracks.Snare, steps(7, 12)) {
		t.Errorf("snare = %v", got.Tracks.Snare)
	}
	if !reflect.DeepEqual(got.Tracks.Kick, p.Tracks.Kick) || !reflect.DeepEqual(got.Tracks.Hat, p.Tracks.Hat) {
		t.Error("kick or hat changed")
	}
}

func TestToggleParity(t *testing.T) {
	got, err := Toggle(readme(), []Edit{{"kick", 2}, {"kick", 2}, {"kick", 2}, {"hat", 1}, {"hat", 1}, {"snare", 16}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Tracks.Kick, steps(0, 1, 4, 8, 12)) {
		t.Errorf("kick = %v", got.Tracks.Kick)
	}
	if !reflect.DeepEqual(got.Tracks.Hat, readme().Tracks.Hat) {
		t.Errorf("hat = %v", got.Tracks.Hat)
	}
	if !reflect.DeepEqual(got.Tracks.Snare, steps(4, 12, 15)) {
		t.Errorf("snare = %v", got.Tracks.Snare)
	}
}

func TestToggleEmpty(t *testing.T) {
	p := readme()
	for _, edits := range [][]Edit{nil, {}} {
		got, err := Toggle(p, edits)
		if err != nil || !reflect.DeepEqual(got, p) {
			t.Fatalf("got %v, %v", got, err)
		}
		got.Tracks.Kick[0] = !got.Tracks.Kick[0]
		if !reflect.DeepEqual(p, readme()) {
			t.Fatal("p shares a track with the copy")
		}
	}
}

func TestToggleCarryOver(t *testing.T) {
	p := readme()
	p.Name, p.Author, p.Swing = "n", "a", 30
	p.ID = json.RawMessage(`"x1"`)
	p.Votes = json.RawMessage(`3`)
	p.CreatedAt = json.RawMessage(`"2020"`)
	got, err := Toggle(p, []Edit{{"kick", 1}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "n" || got.Author != "a" || got.BPM != 120 || got.Swing != 30 ||
		string(got.ID) != `"x1"` || string(got.Votes) != `3` || string(got.CreatedAt) != `"2020"` {
		t.Errorf("not carried over: %+v", got)
	}
	got.ID[1] = 'z'
	got.Votes[0] = '9'
	got.CreatedAt[1] = '9'
	if string(p.ID) != `"x1"` || string(p.Votes) != `3` || string(p.CreatedAt) != `"2020"` {
		t.Error("raw messages shared with p")
	}
	got, _ = Toggle(readme(), nil)
	if got.ID != nil || got.Votes != nil || got.CreatedAt != nil {
		t.Error("absent stored fields must stay nil")
	}
}

func TestToggleNoMutate(t *testing.T) {
	p := readme()
	got, err := Toggle(p, []Edit{{"kick", 1}, {"snare", 2}, {"hat", 3}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p, readme()) {
		t.Fatal("p modified")
	}
	got.Tracks.Kick[5], got.Tracks.Snare[5], got.Tracks.Hat[5] = true, true, true
	if !reflect.DeepEqual(p, readme()) {
		t.Fatal("copy shares tracks with p")
	}
}

func TestToggleErrors(t *testing.T) {
	short := readme()
	short.Tracks.Kick = steps()[:5]
	cases := []struct {
		name  string
		p     Pattern
		edits []Edit
	}{
		{"Snare", readme(), []Edit{{"Snare", 1}}},
		{"tom", readme(), []Edit{{"tom", 1}}},
		{"empty track", readme(), []Edit{{"", 1}}},
		{"step 0", readme(), []Edit{{"kick", 0}}},
		{"step 17", readme(), []Edit{{"kick", 17}}},
		{"step -1", readme(), []Edit{{"kick", -1}}},
		{"bad after good", readme(), []Edit{{"kick", 1}, {"hat", 2}, {"tom", 3}}},
		{"bpm 59", pattern(59, steps(), steps(), steps()), nil},
		{"bpm 59 edits", pattern(59, steps(), steps(), steps()), []Edit{{"kick", 1}}},
		{"short track", short, []Edit{{"kick", 1}}},
	}
	for _, c := range cases {
		got, err := Toggle(c.p, c.edits)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", c.name, err)
		}
		if !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("%s: result not zero", c.name)
		}
	}
}

func TestToggleBoundary(t *testing.T) {
	got, err := Toggle(pattern(120, steps(), steps(), steps()), []Edit{{"kick", 1}, {"kick", 16}})
	if err != nil || !reflect.DeepEqual(got.Tracks.Kick, steps(0, 15)) {
		t.Fatalf("got %v, %v", got.Tracks.Kick, err)
	}
}
