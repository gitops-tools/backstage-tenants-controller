package test

import (
	"regexp"
	"testing"
)

// AssertErrorMatch will fail if the error doesn't match the provided error.
//
// match is treated as a regexp.
func AssertErrorMatch(t *testing.T, match string, e error) {
	t.Helper()
	if !matchErrorString(t, match, e) {
		t.Fatalf("error did not match, got %s, want %s", e, match)
	}
}

func matchErrorString(t *testing.T, match string, e error) bool {
	t.Helper()
	if match == "" && e == nil {
		return true
	}
	if match != "" && e == nil {
		return false
	}
	m, err := regexp.MatchString(match, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return m
}
