package synth

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func trackOf(p Pattern, name string) []bool {
	switch name {
	case "kick":
		return p.Tracks.Kick
	case "snare":
		return p.Tracks.Snare
	}
	return p.Tracks.Hat
}

func randomPattern() Pattern {
	r := rand.New(rand.NewSource(7))
	row := func() []bool {
		s := make([]bool, Steps)
		for i := range s {
			s[i] = r.Intn(2) == 0
		}
		return s
	}
	p := pattern(120, row(), row(), row())
	p.Swing = 20
	p.ID = []byte(`"abc"`)
	p.Votes = []byte(`3`)
	p.CreatedAt = []byte(`"2026-01-01T00:00:00Z"`)
	return p
}

func clonePattern(p Pattern) Pattern {
	c := p
	c.ID = append([]byte(nil), p.ID...)
	c.Votes = append([]byte(nil), p.Votes...)
	c.CreatedAt = append([]byte(nil), p.CreatedAt...)
	c.Tracks = Tracks{
		Kick:  append([]bool(nil), p.Tracks.Kick...),
		Snare: append([]bool(nil), p.Tracks.Snare...),
		Hat:   append([]bool(nil), p.Tracks.Hat...),
	}
	return c
}

func TestRotateReadmeExample(t *testing.T) {
	odd := []int{0, 2, 4, 6, 8, 10, 12, 14}
	p := pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(odd...))
	got, err := Rotate(p, "snare", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Tracks.Snare, steps(6, 14)) {
		t.Errorf("snare = %v", got.Tracks.Snare)
	}
	if !reflect.DeepEqual(got.Tracks.Kick, p.Tracks.Kick) || !reflect.DeepEqual(got.Tracks.Hat, p.Tracks.Hat) {
		t.Error("other tracks changed")
	}
	got, err = Rotate(p, "hat", -1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Tracks.Hat, steps(1, 3, 5, 7, 9, 11, 13, 15)) {
		t.Errorf("hat = %v", got.Tracks.Hat)
	}
	if !reflect.DeepEqual(got.Tracks.Kick, p.Tracks.Kick) || !reflect.DeepEqual(got.Tracks.Snare, p.Tracks.Snare) {
		t.Error("other tracks changed")
	}
}

func TestRotateFormula(t *testing.T) {
	p := randomPattern()
	ns := []int{math.MaxInt, math.MinInt, math.MaxInt - 3, math.MinInt + 5}
	for n := -40; n <= 40; n++ {
		ns = append(ns, n)
	}
	for _, track := range []string{"kick", "snare", "hat"} {
		src := trackOf(p, track)
		for _, n := range ns {
			got, err := Rotate(p, track, n)
			if err != nil {
				t.Fatal(err)
			}
			// reference: big-safe modulo
			mod := n % 16
			if mod < 0 {
				mod += 16
			}
			for i := 1; i <= Steps; i++ {
				want := src[((i-1-mod)%16+16)%16]
				if trackOf(got, track)[i-1] != want {
					t.Fatalf("%s n=%d step %d", track, n, i)
				}
			}
			for _, other := range []string{"kick", "snare", "hat"} {
				if other != track && !reflect.DeepEqual(trackOf(got, other), trackOf(p, other)) {
					t.Fatalf("%s n=%d changed %s", track, n, other)
				}
			}
			again, _ := Rotate(p, track, n%16)
			if !reflect.DeepEqual(got, again) {
				t.Fatalf("%s n=%d differs from n mod 16", track, n)
			}
		}
		zero, _ := Rotate(p, track, 0)
		if !reflect.DeepEqual(zero, p) {
			t.Errorf("%s n=0 differs", track)
		}
		a, _ := Rotate(p, track, 5)
		b, _ := Rotate(p, track, 21)
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s n and n+16 differ", track)
		}
	}
}

func TestRotateOtherFields(t *testing.T) {
	p := randomPattern()
	p.Name, p.Author, p.BPM, p.Swing = "nm", "au-1", 99, 33
	got, err := Rotate(p, "kick", 3)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != p.Name || got.Author != p.Author || got.BPM != p.BPM || got.Swing != p.Swing {
		t.Error("scalar fields differ")
	}
	if string(got.ID) != string(p.ID) || string(got.Votes) != string(p.Votes) || string(got.CreatedAt) != string(p.CreatedAt) {
		t.Error("raw fields differ")
	}
	q := pattern(120, steps(), steps(), steps())
	g, _ := Rotate(q, "hat", 1)
	if g.ID != nil || g.Votes != nil || g.CreatedAt != nil {
		t.Error("absent raw fields should stay nil")
	}
}

func TestRotateNoMutateNoShare(t *testing.T) {
	for _, n := range []int{0, 1, -7, 16} {
		p := randomPattern()
		snap := clonePattern(p)
		got, err := Rotate(p, "snare", n)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p, snap) {
			t.Fatalf("n=%d: p modified", n)
		}
		for _, s := range [][]bool{got.Tracks.Kick, got.Tracks.Snare, got.Tracks.Hat} {
			for i := range s {
				s[i] = !s[i]
			}
		}
		for _, b := range [][]byte{got.ID, got.Votes, got.CreatedAt} {
			for i := range b {
				b[i] = 'X'
			}
		}
		if !reflect.DeepEqual(p, snap) {
			t.Fatalf("n=%d: result shares memory with p", n)
		}
	}
}

func TestRotateErrors(t *testing.T) {
	good := func() Pattern { return pattern(120, steps(0), steps(1), steps(2)) }
	for _, track := range []string{"Hat", "tom", "", "KICK", " kick"} {
		got, err := Rotate(good(), track, 1)
		if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Pattern{}) {
			t.Errorf("track %q: got %v, %v", track, got, err)
		}
	}
	bad := map[string]func(*Pattern){
		"swing": func(p *Pattern) { p.Swing = 61 },
		"bpm":   func(p *Pattern) { p.BPM = 59 },
		"short": func(p *Pattern) { p.Tracks.Hat = p.Tracks.Hat[:15] },
		"auth":  func(p *Pattern) { p.Author = "bad author!" },
	}
	for name, mut := range bad {
		for _, track := range []string{"kick", "tom"} {
			p := good()
			mut(&p)
			got, err := Rotate(p, track, 1)
			if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Pattern{}) {
				t.Errorf("%s/%s: got %v, %v", name, track, got, err)
			}
		}
	}
}
