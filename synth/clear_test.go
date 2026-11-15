package synth

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func readme() Pattern {
	p := pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6, 8, 10, 12, 14))
	p.Name, p.Author = "Four", "me"
	return p
}

func trackOf(p *Pattern, name string) *[]bool {
	switch name {
	case "kick":
		return &p.Tracks.Kick
	case "snare":
		return &p.Tracks.Snare
	}
	return &p.Tracks.Hat
}

func TestClearTracks(t *testing.T) {
	for _, track := range []string{"kick", "snare", "hat"} {
		p := readme()
		got, err := Clear(p, track)
		if err != nil {
			t.Fatal(err)
		}
		want := readme()
		*trackOf(&want, track) = steps()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %+v want %+v", track, got, want)
		}
		if !reflect.DeepEqual(p, readme()) {
			t.Errorf("%s: input modified", track)
		}
	}
}

func TestClearEmptyTrack(t *testing.T) {
	p := pattern(100, steps(), steps(1), steps(2))
	got, err := Clear(p, "kick")
	if err != nil || !reflect.DeepEqual(got, p) {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestClearStoredFields(t *testing.T) {
	p := readme()
	p.ID = json.RawMessage(`7`)
	p.Votes = json.RawMessage(`3`)
	p.CreatedAt = json.RawMessage(`"2024-01-01T00:00:00Z"`)
	got, err := Clear(p, "hat")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.ID) != `7` || string(got.Votes) != `3` || string(got.CreatedAt) != `"2024-01-01T00:00:00Z"` {
		t.Errorf("stored fields changed: %+v", got)
	}
	got.ID[0], got.Votes[0], got.CreatedAt[1] = '9', '9', '9'
	if string(p.ID) != `7` || string(p.Votes) != `3` || string(p.CreatedAt) != `"2024-01-01T00:00:00Z"` {
		t.Errorf("raw message memory shared: %+v", p)
	}

	q := readme()
	g2, err := Clear(q, "kick")
	if err != nil {
		t.Fatal(err)
	}
	if g2.ID != nil || g2.Votes != nil || g2.CreatedAt != nil {
		t.Errorf("absent fields not nil: %+v", g2)
	}
	a, _ := json.Marshal(q)
	b, _ := json.Marshal(Pattern{ID: g2.ID, Name: q.Name, Author: q.Author, BPM: q.BPM, Swing: q.Swing, Tracks: q.Tracks, Votes: g2.Votes, CreatedAt: g2.CreatedAt})
	if string(a) != string(b) {
		t.Errorf("json differs: %s vs %s", a, b)
	}

	var d Pattern
	src := `{"name":"x","author":"me","bpm":90,"swing":0,"tracks":{"kick":[` + rep(16) + `],"snare":[` + rep(16) + `],"hat":[` + rep(16) + `]}}`
	if err := json.Unmarshal([]byte(src), &d); err != nil {
		t.Fatal(err)
	}
	g3, err := Clear(d, "snare")
	if err != nil || g3.Votes != nil || g3.ID != nil || g3.CreatedAt != nil {
		t.Errorf("decoded: %+v, %v", g3, err)
	}
}

func rep(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			s += ","
		}
		s += "true"
	}
	return s
}

func TestClearNoSharing(t *testing.T) {
	for _, track := range []string{"kick", "snare", "hat"} {
		p := readme()
		got, err := Clear(p, track)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"kick", "snare", "hat"} {
			gs := *trackOf(&got, name)
			ps := *trackOf(&p, name)
			gs[3] = !gs[3]
			if !reflect.DeepEqual(p, readme()) {
				t.Errorf("%s/%s: result mutation changed input", track, name)
			}
			gs[3] = !gs[3]
			snapshot := append([]bool(nil), gs...)
			ps[3] = !ps[3]
			if !reflect.DeepEqual(gs, snapshot) {
				t.Errorf("%s/%s: input mutation changed result", track, name)
			}
			ps[3] = !ps[3]
		}
	}
}

func TestClearErrors(t *testing.T) {
	for _, track := range []string{"Kick", "clap", "", " kick"} {
		got, err := Clear(readme(), track)
		if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("track %q: %+v, %v", track, got, err)
		}
	}
	bad := map[string]func(*Pattern){
		"author": func(p *Pattern) { p.Author = "" },
		"bpm":    func(p *Pattern) { p.BPM = 59 },
		"swing":  func(p *Pattern) { p.Swing = 61 },
		"short":  func(p *Pattern) { p.Tracks.Hat = p.Tracks.Hat[:15] },
	}
	for name, mutate := range bad {
		p := readme()
		mutate(&p)
		got, err := Clear(p, "snare")
		if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("%s: %+v, %v", name, got, err)
		}
	}
}

func TestClearPartialStoredFields(t *testing.T) {
	p := readme()
	p.ID = json.RawMessage(`7`)
	got, err := Clear(p, "kick")
	if err != nil {
		t.Fatal(err)
	}
	if got.Votes != nil || got.CreatedAt != nil || string(got.ID) != `7` {
		t.Errorf("stored fields: %+v", got)
	}
}
