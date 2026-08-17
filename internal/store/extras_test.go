package store

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestCountAndSearchDomainHelpers(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	payload := json.RawMessage(`{"eva":"8000105","name":"Frankfurt Hbf BLS"}`)
	if _, _, err := s.UpsertStation([]json.RawMessage{payload}); err != nil {
		t.Fatal(err)
	}
	counts, err := s.CountByResourceType()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range counts {
		if row.ResourceType == "station" && row.Count >= 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected station count, got %#v", counts)
	}
	hits, err := s.SearchStation("Frankfurt", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected SearchStation hit")
	}
}
