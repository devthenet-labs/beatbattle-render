package synth

import (
	"math"
)

// SampleRate is the output rate in Hz.
const SampleRate = 44100

const (
	kickDur  = 0.120
	snareDur = 0.100
	hatDur   = 0.040

	// decayRate makes an envelope fall to e^-6 (about -52 dB) at its end.
	decayRate = 6.0

	// Gains are chosen so that the loudest possible coincidence (kick, snare
	// and hat on one step) peaks below 0.8 of full scale before the master
	// gain, and the master gain leaves headroom for overlapping tails at
	// high tempos.
	kickGain       = 0.60
	snareNoiseGain = 0.22
	snareToneGain  = 0.18
	hatGain        = 0.14
	masterGain     = 0.85

	hatPole = 0.85 // one-pole high-pass coefficient
	lcgSeed = 0x2545F491
)

// lcg is a 32-bit linear congruential generator (Numerical Recipes
// constants) used for all noise, so renders are reproducible.
type lcg uint32

// next returns a value in [-1, 1).
func (l *lcg) next() float64 {
	*l = *l*1664525 + 1013904223
	return float64(int32(*l)) / (1 << 31)
}

// stepStart returns the first sample of step i on the swung sixteenth-note
// grid, rounded to the nearest sample. Even steps sit on the straight grid
// (60/bpm/4 seconds per step); each odd step (the off-beats) is pushed
// later by swing percent of one step's duration, towards the next step.
// swing is 0-100.
func stepStart(i, bpm, swing int) int {
	const perBeat = SampleRate * 15 // samples*bpm per sixteenth
	start := (i*perBeat + bpm/2) / bpm
	if i%2 == 1 {
		start += (swing*perBeat + 50*bpm) / (100 * bpm)
	}
	return start
}

// Length returns the sample count Render produces for a tempo.
func Length(bpm int) int { return stepStart(Steps, bpm, 0) }

// Render validates p and renders it. The result has Length(p.BPM) samples;
// voices still ringing at the end of the last step are cut. Off-beat
// (odd-index) steps are delayed by p.Swing percent of a step; see
// stepStart.
func Render(p Pattern) ([]int16, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	n := Length(p.BPM)
	mix := make([]float64, n)
	noise := lcg(lcgSeed)
	for i := 0; i < Steps; i++ {
		start := stepStart(i, p.BPM, p.Swing)
		if p.Tracks.Kick[i] {
			addKick(mix, start)
		}
		if p.Tracks.Snare[i] {
			addSnare(mix, start, &noise)
		}
		if p.Tracks.Hat[i] {
			addHat(mix, start, &noise)
		}
	}
	out := make([]int16, n)
	for i, v := range mix {
		out[i] = clip(v * masterGain)
	}
	return out, nil
}

func clip(v float64) int16 {
	s := math.Round(v * math.MaxInt16)
	if s > math.MaxInt16 {
		return math.MaxInt16
	}
	if s < math.MinInt16 {
		return math.MinInt16
	}
	return int16(s)
}

// voiceLen is the number of samples a voice of dur seconds spans, bounded by
// the room left in the buffer.
func voiceLen(mix []float64, start int, dur float64) int {
	return min(int(dur*SampleRate), len(mix)-start)
}

func envelope(t, dur float64) float64 { return math.Exp(-decayRate * t / dur) }

// addKick adds a sine sweeping exponentially 150 Hz to 50 Hz over its length.
func addKick(mix []float64, start int) {
	phase := 0.0
	for k := 0; k < voiceLen(mix, start, kickDur); k++ {
		t := float64(k) / SampleRate
		freq := 150 * math.Pow(50.0/150.0, t/kickDur)
		phase += 2 * math.Pi * freq / SampleRate
		mix[start+k] += kickGain * envelope(t, kickDur) * math.Sin(phase)
	}
}

// addSnare adds LCG noise plus a 200 Hz sine.
func addSnare(mix []float64, start int, noise *lcg) {
	for k := 0; k < voiceLen(mix, start, snareDur); k++ {
		t := float64(k) / SampleRate
		tone := math.Sin(2 * math.Pi * 200 * t)
		mix[start+k] += envelope(t, snareDur) * (snareNoiseGain*noise.next() + snareToneGain*tone)
	}
}

// addHat adds LCG noise through a one-pole high-pass.
func addHat(mix []float64, start int, noise *lcg) {
	var prevIn, prevOut float64
	for k := 0; k < voiceLen(mix, start, hatDur); k++ {
		t := float64(k) / SampleRate
		in := noise.next()
		out := hatPole * (prevOut + in - prevIn)
		prevIn, prevOut = in, out
		mix[start+k] += hatGain * envelope(t, hatDur) * out
	}
}
