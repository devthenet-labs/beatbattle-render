package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/devthenet-labs/beatbattle-render/synth"
)

// maxBody bounds every request body.
const maxBody = 16 << 10

// newHandler serves the render service under /render: readiness at
// /render/healthz and the stateless POST /render/wav.
func newHandler(revision string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/render/healthz", allow(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "revision": revision})
	}))
	mux.HandleFunc("/render/wav", allow(http.MethodPost, handleWAV))
	mux.HandleFunc("/render/analyze", allow(http.MethodPost, handleAnalyze))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	return withSecurityHeaders(mux)
}

// allow restricts a handler to one method, answering the rest with a JSON 405.
func allow(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next(w, r)
	}
}

func handleWAV(w http.ResponseWriter, r *http.Request) {
	p, err := decodePattern(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	samples, err := synth.Render(p)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	wav := synth.EncodeWAV(samples)
	h := w.Header()
	h.Set("Content-Type", "audio/wav")
	h.Set("Content-Disposition", `inline; filename="beat.wav"`)
	h.Set("Content-Length", strconv.Itoa(len(wav)))
	_, _ = w.Write(wav)
}

// analyzeResponse is the POST /render/analyze response.
type analyzeResponse struct {
	DurationMs int64   `json:"durationMs"`
	Onsets     []int64 `json:"onsets"`
	Steps      []int64 `json:"steps"`
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	p, err := decodePattern(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a, err := synth.Analyze(p)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, analyzeResponse{DurationMs: a.DurationMs, Onsets: a.OnsetsMs, Steps: a.StepsMs})
}

// decodePattern reads exactly one pattern object, rejecting unknown fields
// and trailing data.
func decodePattern(body io.Reader) (synth.Pattern, error) {
	var p synth.Pattern
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return p, err
		}
		return p, errors.New("invalid pattern: body is not a valid pattern JSON object: " + err.Error())
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return p, err
		}
		return p, errors.New("invalid pattern: unexpected data after the JSON object")
	}
	return p, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// withSecurityHeaders sets the headers every response carries.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
