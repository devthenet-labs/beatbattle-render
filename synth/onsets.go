package synth

import (
	"math"
	"time"
)

const (
	onsetWindow    = 128  // samples per RMS window
	onsetHop       = 32   // samples between windows
	onsetFloor     = 50.0 // absolute RMS below which nothing is an onset
	onsetThreshold = 0.1  // fraction of the loudest window's RMS
)

// Onsets returns the times at which the windowed RMS envelope of samples
// rises past a threshold (a tenth of its peak, and never below a small
// absolute floor). The detector re-arms once the envelope falls below half
// the threshold. Each onset is the start of the first window that crossed.
func Onsets(samples []int16, sampleRate int) []time.Duration {
	env := envelope16(samples)
	peak := 0.0
	for _, v := range env {
		peak = math.Max(peak, v)
	}
	if peak < onsetFloor {
		return nil
	}
	threshold := math.Max(onsetThreshold*peak, onsetFloor)
	var out []time.Duration
	armed := true
	for k, v := range env {
		switch {
		case armed && v >= threshold:
			out = append(out, time.Duration(k*onsetHop)*time.Second/time.Duration(sampleRate))
			armed = false
		case !armed && v < threshold/2:
			armed = true
		}
	}
	return out
}

// envelope16 returns the RMS of each window, hopping by onsetHop. Windows
// that run past the end see the missing samples as silence.
func envelope16(samples []int16) []float64 {
	var env []float64
	for start := 0; start < len(samples); start += onsetHop {
		end := min(start+onsetWindow, len(samples))
		var sum float64
		for _, s := range samples[start:end] {
			sum += float64(s) * float64(s)
		}
		env = append(env, math.Sqrt(sum/onsetWindow))
	}
	return env
}
