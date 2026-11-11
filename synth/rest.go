package synth

// LongestRest returns the length, in steps, of the longest run of consecutive
// steps on which the named track ("kick", "snare" or "hat") does not play.
// There is no wrap-around from step 16 to step 1. It returns -1 for any other
// track name and for a pattern that Validate refuses.
func LongestRest(p Pattern, track string) int {
	if !restPlayable(p) {
		return -1
	}
	var row []bool
	switch track {
	case "kick":
		row = p.Tracks.Kick
	case "snare":
		row = p.Tracks.Snare
	case "hat":
		row = p.Tracks.Hat
	default:
		return -1
	}
	longest, run := 0, 0
	for _, on := range row {
		if on {
			run = 0
			continue
		}
		run++
		if run > longest {
			longest = run
		}
	}
	return longest
}

// restPlayable reports whether p has a valid tempo, swing and step grid.
func restPlayable(p Pattern) bool {
	if p.BPM < 60 || p.BPM > 200 || p.Swing < 0 || p.Swing > 60 {
		return false
	}
	return len(p.Tracks.Kick) == Steps && len(p.Tracks.Snare) == Steps && len(p.Tracks.Hat) == Steps
}
