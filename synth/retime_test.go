package synth

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"
)

func retimeFixture() Pattern {
	odd := make([]bool, Steps)
	for i := 0; i < Steps; i += 2 {
		odd[i] = true
	}
	on := func(idx ...int) []bool {
		s := make([]bool, Steps)
		for _, i := range idx {
			s[i-1] = true
		}
		return s
	}
	return Pattern{Name: "Four", Author: "me", BPM: 120, Swing: 10,
		Tracks: Tracks{Kick: on(1, 5, 9, 13), Snare: on(5, 13), Hat: odd}}
}

func TestRetimeReadmeExample(t *testing.T) {
	p := retimeFixture()
	got, err := Retime(p, 90, 25)
	if err != nil {
		t.Fatal(err)
	}
	if got.BPM != 90 || got.Swing != 25 || got.Name != "Four" || got.Author != "me" {
		t.Fatalf("got %+v", got)
	}
	if !slices.Equal(got.Tracks.Kick, p.Tracks.Kick) || !slices.Equal(got.Tracks.Snare, p.Tracks.Snare) || !slices.Equal(got.Tracks.Hat, p.Tracks.Hat) {
		t.Fatal("tracks differ")
	}
	if p.BPM != 120 || p.Swing != 10 {
		t.Fatal("p modified")
	}
}

func TestRetimeBoundaries(t *testing.T) {
	for _, c := range [][2]int{{60, 0}, {200, 60}, {60, 60}, {200, 0}} {
		if _, err := Retime(retimeFixture(), c[0], c[1]); err != nil {
			t.Errorf("%v: %v", c, err)
		}
	}
}

func TestRetimeRejects(t *testing.T) {
	short := retimeFixture()
	short.Tracks.Kick = short.Tracks.Kick[:15]
	noAuthor := retimeFixture()
	noAuthor.Author = ""
	for name, c := range map[string]struct {
		p          Pattern
		bpm, swing int
	}{
		"bpm59":    {retimeFixture(), 59, 10},
		"bpm201":   {retimeFixture(), 201, 10},
		"swing-1":  {retimeFixture(), 100, -1},
		"swing61":  {retimeFixture(), 100, 61},
		"author":   {noAuthor, 100, 10},
		"15steps":  {short, 100, 10},
		"zeroPatt": {Pattern{}, 100, 10},
	} {
		got, err := Retime(c.p, c.bpm, c.swing)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
		if !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("%s: not zero", name)
		}
	}
}

func TestRetimeStoredFields(t *testing.T) {
	p := retimeFixture()
	got, err := Retime(p, 90, 25)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != nil || got.Votes != nil || got.CreatedAt != nil {
		t.Fatal("absent fields should stay nil")
	}

	p.ID = json.RawMessage("7")
	p.Votes = json.RawMessage("3")
	p.CreatedAt = json.RawMessage(`"2024-01-02T03:04:05Z"`)
	got, err = Retime(p, 90, 25)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.ID, p.ID) || !bytes.Equal(got.Votes, p.Votes) || !bytes.Equal(got.CreatedAt, p.CreatedAt) {
		t.Fatal("stored fields differ")
	}
	var a, b map[string]any
	pj, _ := json.Marshal(p)
	gj, _ := json.Marshal(got)
	if err := json.Unmarshal(pj, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(gj, &b); err != nil {
		t.Fatal(err)
	}
	a["bpm"], a["swing"] = float64(90), float64(25)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("%s vs %s", pj, gj)
	}

	// empty non-nil stays non-nil
	p.ID = json.RawMessage{}
	got, _ = Retime(p, 90, 25)
	if got.ID == nil {
		t.Fatal("empty ID became nil")
	}
}

func TestRetimeDecodedWithoutID(t *testing.T) {
	var p Pattern
	src := `{"name":"Four","author":"me","bpm":120,"swing":0,"tracks":{"kick":[true,false,false,false,false,false,false,false,false,false,false,false,false,false,false,false],"snare":[false,false,false,false,false,false,false,false,false,false,false,false,false,false,false,false],"hat":[false,false,false,false,false,false,false,false,false,false,false,false,false,false,false,false]}}`
	if err := json.Unmarshal([]byte(src), &p); err != nil {
		t.Fatal(err)
	}
	got, err := Retime(p, 100, 5)
	if err != nil {
		t.Fatal(err)
	}
	pj, _ := json.Marshal(p)
	gj, _ := json.Marshal(got)
	var a, b map[string]json.RawMessage
	_ = json.Unmarshal(pj, &a)
	_ = json.Unmarshal(gj, &b)
	if !bytes.Equal(a["id"], b["id"]) {
		t.Fatalf("id %s vs %s", a["id"], b["id"])
	}
}

func TestRetimeNoAliasing(t *testing.T) {
	p := retimeFixture()
	p.ID = json.RawMessage("7")
	snap := retimeFixture()
	snap.ID = json.RawMessage("7")
	got, err := Retime(p, 90, 25)
	if err != nil {
		t.Fatal(err)
	}
	if p.BPM != 120 || !reflect.DeepEqual(p, snap) {
		t.Fatal("p mutated")
	}
	want := retimeFixture()
	got.Tracks.Kick[0] = !got.Tracks.Kick[0]
	got.Tracks.Snare[0] = !got.Tracks.Snare[0]
	got.Tracks.Hat[0] = !got.Tracks.Hat[0]
	got.ID[0] = '9'
	if !reflect.DeepEqual(p.Tracks, want.Tracks) || string(p.ID) != "7" {
		t.Fatal("result aliases p")
	}
	got, _ = Retime(p, 90, 25)
	p.Tracks.Kick[1] = true
	p.Tracks.Snare[1] = true
	p.Tracks.Hat[1] = true
	if !reflect.DeepEqual(got.Tracks, want.Tracks) {
		t.Fatal("p change leaked into result")
	}
}
