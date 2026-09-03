package engine

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeErrorPrefersLastJSONError(t *testing.T) {
	stderr := []byte(strings.Join([]string{
		`{"type":"progress","phase":"fetching","name":"Chill","index":1,"total":2}`,
		`{"error":"MusicKit request timed out"}`,
	}, "\n"))

	err := decodeError(stderr, errors.New("exit status 1"))
	if got := err.Error(); got != "MusicKit request timed out" {
		t.Fatalf("got %q", got)
	}
}

func TestDecodeErrorFallsBackToWaitErr(t *testing.T) {
	wait := errors.New("exit status 1")
	err := decodeError(nil, wait)
	if !errors.Is(err, wait) {
		t.Fatalf("got %v want %v", err, wait)
	}
}
