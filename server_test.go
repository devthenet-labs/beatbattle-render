package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const kickPattern = `{"name":"Four","author":"brvtl","bpm":120,"swing":0,"tracks":{
"kick":[true,false,false,false,true,false,false,false,true,false,false,false,true,false,false,false],
"snare":[false,false,false,false,false,false,false,false,false,false,false,false,false,false,false,false],
"hat":[false,false,false,false,false,false,false,false,false,false,false,false,false,false,false,false]}}`

func TestHandler(t *testing.T) {
	handler := newHandler("abc123")
	cases := []struct {
		name, method, path, body string
		status                   int
		contentType, bodyHas     string
	}{
		{"readiness", http.MethodGet, "/render/healthz", "", 200, "application/json", `"status":"ok"`},
		{"root is not served", http.MethodGet, "/", "", 404, "application/json", `"error"`},
		{"unknown path", http.MethodGet, "/render/missing", "", 404, "application/json", `"error"`},
		{"no prefix healthz", http.MethodGet, "/healthz", "", 404, "application/json", `"error"`},
		{"wav wrong method", http.MethodGet, "/render/wav", "", 405, "application/json", `"error"`},
		{"wav ok", http.MethodPost, "/render/wav", kickPattern, 200, "audio/wav", "RIFF"},
		{"wav not json", http.MethodPost, "/render/wav", "nope", 400, "application/json", `"error"`},
		{"wav empty", http.MethodPost, "/render/wav", "", 400, "application/json", `"error"`},
		{"wav unknown field", http.MethodPost, "/render/wav", strings.Replace(kickPattern, `"swing":0`, `"swing":0,"extra":1`, 1), 400, "application/json", `"error"`},
		{"wav unknown track", http.MethodPost, "/render/wav", strings.Replace(kickPattern, `"hat"`, `"cowbell"`, 1), 400, "application/json", `"error"`},
		{"wav fractional bpm", http.MethodPost, "/render/wav", strings.Replace(kickPattern, `"bpm":120`, `"bpm":120.5`, 1), 400, "application/json", `"error"`},
		{"wav bpm out of range", http.MethodPost, "/render/wav", strings.Replace(kickPattern, `"bpm":120`, `"bpm":300`, 1), 400, "application/json", `"error"`},
		{"wav trailing data", http.MethodPost, "/render/wav", kickPattern + `{}`, 400, "application/json", `"error"`},
		{"wav ignores server fields", http.MethodPost, "/render/wav", strings.Replace(kickPattern, `"swing":0`, `"id":"p_1","votes":3,"createdAt":"2026-10-26T21:00:00Z"`, 1), 200, "audio/wav", "RIFF"},
		{"wav too large", http.MethodPost, "/render/wav", `{"name":"` + strings.Repeat("a", 20<<10) + `"}`, 413, "application/json", `"error"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if rec.Code != tc.status {
				t.Fatalf("%s %s = %d, want %d: %s", tc.method, tc.path, rec.Code, tc.status, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != tc.contentType {
				t.Errorf("Content-Type = %q, want %q", got, tc.contentType)
			}
			if !strings.Contains(rec.Body.String(), tc.bodyHas) {
				t.Errorf("body lacks %q", tc.bodyHas)
			}
			if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
		})
	}
}

func TestReadinessReportsRevision(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler("abc123").ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/render/healthz", nil))
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("readiness body is not JSON: %v", err)
	}
	if got["revision"] != "abc123" {
		t.Errorf("revision = %q, want abc123", got["revision"])
	}
}

func TestWAVResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler("x").ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/render/wav", strings.NewReader(kickPattern)))
	if got := rec.Header().Get("Content-Disposition"); got != `inline; filename="beat.wav"` {
		t.Errorf("Content-Disposition = %q", got)
	}
	if n := rec.Body.Len(); n != 176444 {
		t.Errorf("body = %d bytes, want 176444", n)
	}
}
