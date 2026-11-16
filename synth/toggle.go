package synth

import "encoding/json"

// Edit names one step of one track to flip. Step is 1-based (1 to Steps).
type Edit struct {
	Track string
	Step  int
}

// Toggle returns a copy of p in which each edit, in order, flips its step of
// its track. p is not modified and the copy shares no memory with it. It
// returns the zero Pattern and an error wrapping ErrInvalid if p is invalid or
// any edit names an unknown track or a step outside 1 to Steps.
func Toggle(p Pattern, edits []Edit) (Pattern, error) {
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	for _, e := range edits {
		switch e.Track {
		case "kick", "snare", "hat":
		default:
			return Pattern{}, invalid("unknown track %q", e.Track)
		}
		if e.Step < 1 || e.Step > Steps {
			return Pattern{}, invalid("step must be an integer from 1 to %d", Steps)
		}
	}

	out := p
	out.ID = cloneRaw(p.ID)
	out.Votes = cloneRaw(p.Votes)
	out.CreatedAt = cloneRaw(p.CreatedAt)
	out.Tracks = Tracks{
		Kick:  append([]bool(nil), p.Tracks.Kick...),
		Snare: append([]bool(nil), p.Tracks.Snare...),
		Hat:   append([]bool(nil), p.Tracks.Hat...),
	}
	for _, e := range edits {
		var row []bool
		switch e.Track {
		case "kick":
			row = out.Tracks.Kick
		case "snare":
			row = out.Tracks.Snare
		case "hat":
			row = out.Tracks.Hat
		}
		row[e.Step-1] = !row[e.Step-1]
	}
	return out, nil
}

func cloneRaw(m json.RawMessage) json.RawMessage {
	if m == nil {
		return nil
	}
	return append(json.RawMessage{}, m...)
}
