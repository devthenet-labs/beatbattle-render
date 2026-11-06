package synth

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("valid pattern round-trips", func(t *testing.T) {
		want := pattern(120, steps(0, 4, 8, 12), steps(4, 12), steps(0, 2, 4, 6))
		want.Name = "Four"
		want.Author = "tester"
		want.Swing = 25
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse() error = %v, want nil", err)
		}
		if got.Name != want.Name || got.Author != want.Author || got.BPM != want.BPM || got.Swing != want.Swing {
			t.Fatalf("Parse() = %+v, want %+v", got, want)
		}
		for _, trk := range []struct {
			name      string
			got, want []bool
		}{
			{"kick", got.Tracks.Kick, want.Tracks.Kick},
			{"snare", got.Tracks.Snare, want.Tracks.Snare},
			{"hat", got.Tracks.Hat, want.Tracks.Hat},
		} {
			if len(trk.got) != len(trk.want) {
				t.Fatalf("%s length = %d, want %d", trk.name, len(trk.got), len(trk.want))
			}
			for i := range trk.want {
				if trk.got[i] != trk.want[i] {
					t.Fatalf("%s[%d] = %v, want %v", trk.name, i, trk.got[i], trk.want[i])
				}
			}
		}
	})

	t.Run("unknown field refused", func(t *testing.T) {
		p := pattern(120, steps(), steps(), steps())
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		withExtra := strings.Replace(string(data), "{", `{"extra":true,`, 1)
		if _, err := Parse([]byte(withExtra)); err == nil {
			t.Fatal("Parse() with unknown field succeeded, want error")
		}
	})

	t.Run("malformed JSON refused", func(t *testing.T) {
		if _, err := Parse([]byte("not json")); err == nil {
			t.Fatal("Parse() of malformed JSON succeeded, want error")
		}
	})

	t.Run("trailing data refused", func(t *testing.T) {
		p := pattern(120, steps(), steps(), steps())
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		withTrailing := append(data, []byte(" garbage")...)
		if _, err := Parse(withTrailing); err == nil {
			t.Fatal("Parse() with trailing data succeeded, want error")
		}
	})

	t.Run("decodes but fails validation", func(t *testing.T) {
		p := pattern(999, steps(), steps(), steps())
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		_, err = Parse(data)
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Parse() error = %v, want errors.Is(err, ErrInvalid)", err)
		}
	})

	t.Run("over the length limit refused without decoding", func(t *testing.T) {
		p := pattern(120, steps(), steps(), steps())
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		padded := append(bytes.Repeat([]byte(" "), 16385-len(data)), data...)
		if len(padded) != 16385 {
			t.Fatalf("len(padded) = %d, want 16385", len(padded))
		}
		_, err = Parse(padded)
		if !errors.Is(err, ErrTooLong) {
			t.Fatalf("Parse() error = %v, want errors.Is(err, ErrTooLong)", err)
		}
	})

	t.Run("exactly 16384 bytes not refused for length", func(t *testing.T) {
		p := pattern(120, steps(), steps(), steps())
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		padded := append(bytes.Repeat([]byte(" "), 16384-len(data)), data...)
		if len(padded) != 16384 {
			t.Fatalf("len(padded) = %d, want 16384", len(padded))
		}
		_, err = Parse(padded)
		if errors.Is(err, ErrTooLong) {
			t.Fatalf("Parse() error = %v, want no ErrTooLong", err)
		}
	})
}
