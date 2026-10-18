package state

import (
	"path/filepath"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsNew("a") {
		t.Error("a should be new")
	}
	if s.IsNew("a") {
		t.Error("a should not be new twice")
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s2.IsNew("a") {
		t.Error("a should persist as seen")
	}
	if !s2.IsNew("b") {
		t.Error("b should be new")
	}
}
