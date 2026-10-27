package synth

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func steps(on ...int) []bool {
	s := make([]bool, Steps)
	for _, i := range on {
		s[i] = true
	}
	return s
}

func pattern(bpm int, kick, snare, hat []bool) Pattern {
	return Pattern{Name: "t", Author: "tester", BPM: bpm, Tracks: Tracks{Kick: kick, Snare: snare, Hat: hat}}
}

func TestValidate(t *testing.T) {
	good := pattern(120, steps(), steps(), steps())
	cases := []struct {
		name   string
		mutate func(*Pattern)
		ok     bool
	}{
		{"valid", func(*Pattern) {}, true},
		{"name trimmed to empty", func(p *Pattern) { p.Name = "   " }, false},
		{"name 40 runes", func(p *Pattern) { p.Name = strings.Repeat("é", 40) }, true},
		{"name 41 runes", func(p *Pattern) { p.Name = strings.Repeat("é", 41) }, false},
		{"name control char", func(p *Pattern) { p.Name = "a\tb" }, false},
		{"author empty", func(p *Pattern) { p.Author = "" }, false},
		{"author 24", func(p *Pattern) { p.Author = strings.Repeat("a", 24) }, true},
		{"author 25", func(p *Pattern) { p.Author = strings.Repeat("a", 25) }, false},
		{"author charset", func(p *Pattern) { p.Author = "a_b-9Z" }, true},
		{"author space", func(p *Pattern) { p.Author = "a b" }, false},
		{"author unicode", func(p *Pattern) { p.Author = "é" }, false},
		{"bpm 59", func(p *Pattern) { p.BPM = 59 }, false},
		{"bpm 60", func(p *Pattern) { p.BPM = 60 }, true},
		{"bpm 200", func(p *Pattern) { p.BPM = 200 }, true},
		{"bpm 201", func(p *Pattern) { p.BPM = 201 }, false},
		{"swing -1", func(p *Pattern) { p.Swing = -1 }, false},
		{"swing 60", func(p *Pattern) { p.Swing = 60 }, true},
		{"swing 61", func(p *Pattern) { p.Swing = 61 }, false},
		{"short kick", func(p *Pattern) { p.Tracks.Kick = make([]bool, 15) }, false},
		{"long snare", func(p *Pattern) { p.Tracks.Snare = make([]bool, 17) }, false},
		{"missing hat", func(p *Pattern) { p.Tracks.Hat = nil }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := good
			tc.mutate(&p)
			err := p.Validate()
			if (err == nil) != tc.ok {
				t.Fatalf("Validate() = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

func TestRenderRejectsInvalid(t *testing.T) {
	if _, err := Render(Pattern{}); err == nil {
		t.Fatal("Render(zero pattern) succeeded")
	}
}

func TestWAVHeaderAndLength(t *testing.T) {
	samples, err := Render(pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6)))
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 88200 {
		t.Fatalf("samples = %d, want 88200", len(samples))
	}
	wav := EncodeWAV(samples)
	if len(wav) != 44+88200*2 {
		t.Fatalf("wav = %d bytes, want %d", len(wav), 44+88200*2)
	}
	le := binary.LittleEndian
	checks := []struct {
		name      string
		got, want any
	}{
		{"RIFF", string(wav[0:4]), "RIFF"},
		{"riff size", le.Uint32(wav[4:]), uint32(len(wav) - 8)},
		{"WAVE", string(wav[8:12]), "WAVE"},
		{"fmt ", string(wav[12:16]), "fmt "},
		{"fmt size", le.Uint32(wav[16:]), uint32(16)},
		{"format", le.Uint16(wav[20:]), uint16(1)},
		{"channels", le.Uint16(wav[22:]), uint16(1)},
		{"rate", le.Uint32(wav[24:]), uint32(44100)},
		{"byte rate", le.Uint32(wav[28:]), uint32(88200)},
		{"block align", le.Uint16(wav[32:]), uint16(2)},
		{"bits", le.Uint16(wav[34:]), uint16(16)},
		{"data", string(wav[36:40]), "data"},
		{"data size", le.Uint32(wav[40:]), uint32(88200 * 2)},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestDeterministic(t *testing.T) {
	p := pattern(137, steps(0, 3, 8), steps(4, 12, 13), steps(1, 2, 3, 5, 7, 9))
	a, _ := Render(p)
	b, _ := Render(p)
	if !bytes.Equal(EncodeWAV(a), EncodeWAV(b)) {
		t.Fatal("two renders differ")
	}
}

func TestEmptyPatternIsSilent(t *testing.T) {
	samples, err := Render(pattern(90, steps(), steps(), steps()))
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range samples {
		if s != 0 {
			t.Fatalf("sample %d = %d, want silence", i, s)
		}
	}
	if got := Onsets(samples, SampleRate); len(got) != 0 {
		t.Fatalf("onsets in silence: %v", got)
	}
}

func TestFullPatternDoesNotClip(t *testing.T) {
	all := steps(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	for _, bpm := range []int{60, 120, 200} {
		samples, err := Render(pattern(bpm, all, all, all))
		if err != nil {
			t.Fatal(err)
		}
		peak := 0
		for _, s := range samples {
			peak = max(peak, abs(int(s)))
		}
		if peak >= 32767 || peak < 8000 {
			t.Errorf("bpm %d: peak = %d, want loud (>=8000) but unclipped (<32767)", bpm, peak)
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func TestKickOnsets(t *testing.T) {
	samples, err := Render(pattern(120, steps(0, 4, 8, 12), steps(), steps()))
	if err != nil {
		t.Fatal(err)
	}
	got := Onsets(samples, SampleRate)
	want := []time.Duration{0, 500 * time.Millisecond, 1000 * time.Millisecond, 1500 * time.Millisecond}
	if len(got) != len(want) {
		t.Fatalf("onsets = %v, want %d of them", got, len(want))
	}
	for i := range want {
		if d := got[i] - want[i]; d < -5*time.Millisecond || d > 5*time.Millisecond {
			t.Errorf("onset %d = %v, want %v ±5ms", i, got[i], want[i])
		}
	}
}

func TestSnareAndHatOnsets(t *testing.T) {
	for name, p := range map[string]Pattern{
		"snare": pattern(120, steps(), steps(4, 12), steps()),
		"hat":   pattern(120, steps(), steps(), steps(4, 12)),
	} {
		samples, _ := Render(p)
		got := Onsets(samples, SampleRate)
		if len(got) != 2 {
			t.Fatalf("%s onsets = %v, want 2", name, got)
		}
		for i, want := range []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond} {
			if d := got[i] - want; d < -5*time.Millisecond || d > 5*time.Millisecond {
				t.Errorf("%s onset %d = %v, want %v ±5ms", name, i, got[i], want)
			}
		}
	}
}

func randomSteps(r *rand.Rand) []bool {
	s := make([]bool, Steps)
	for i := range s {
		s[i] = r.Intn(2) == 0
	}
	return s
}

func TestSwingDelaysOffBeatSteps(t *testing.T) {
	for _, swing := range []int{0, 25, 60} {
		p := pattern(120, steps(1, 3), steps(), steps())
		p.Swing = swing
		samples, err := Render(p)
		if err != nil {
			t.Fatalf("swing %d: %v", swing, err)
		}
		got := Onsets(samples, SampleRate)
		if len(got) != 2 {
			t.Fatalf("swing %d: onsets = %v, want 2", swing, got)
		}
		for i, step := range []int{1, 3} {
			want := time.Duration(step)*125*time.Millisecond + time.Duration(swing)*125*time.Millisecond/100
			if d := got[i] - want; d < -5*time.Millisecond || d > 5*time.Millisecond {
				t.Errorf("swing %d: onset %d = %v, want %v ±5ms", swing, i, got[i], want)
			}
		}
	}
}

func TestSwingNeverChangesLength(t *testing.T) {
	all := steps(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	for _, bpm := range []int{60, 120, 200} {
		for _, swing := range []int{0, 30, 60} {
			p := pattern(bpm, all, all, all)
			p.Swing = swing
			samples, err := Render(p)
			if err != nil {
				t.Fatalf("bpm %d swing %d: %v", bpm, swing, err)
			}
			if len(samples) != Length(bpm) {
				t.Errorf("bpm %d swing %d: len = %d, want %d", bpm, swing, len(samples), Length(bpm))
			}
		}
	}
}

func TestSwingZeroMatchesStraightGrid(t *testing.T) {
	all := steps(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	a := pattern(120, all, all, all)
	a.Swing = 0
	b := pattern(120, all, all, all)
	samplesA, err := Render(a)
	if err != nil {
		t.Fatal(err)
	}
	samplesB, err := Render(b)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(EncodeWAV(samplesA), EncodeWAV(samplesB)) {
		t.Fatal("swing=0 differs from swing left at zero value")
	}
}

func TestRandomPatternsProperty(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		p := pattern(60+r.Intn(141), randomSteps(r), randomSteps(r), randomSteps(r))
		p.Swing = r.Intn(61)
		samples, err := Render(p)
		if err != nil {
			t.Fatalf("case %d (%+v): %v", i, p, err)
		}
		// 16 sixteenth notes of 60/bpm/4 s: 44100*240/bpm samples, to the nearest.
		want := (44100*240 + p.BPM/2) / p.BPM
		if len(samples) != want {
			t.Fatalf("case %d bpm %d: %d samples, want %d", i, p.BPM, len(samples), want)
		}
		if wav := EncodeWAV(samples); len(wav) != 44+2*want {
			t.Fatalf("case %d: wav %d bytes", i, len(wav))
		}
	}
}
