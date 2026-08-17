package cliutil

import (
	"context"
	"testing"
	"time"
)

func TestEnsureFreshNilDB(t *testing.T) {
	d, err := EnsureFresh(context.Background(), nil, []string{"station"}, Policy{StaleAfter: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if d != DecisionNoStore {
		t.Fatalf("got %s want no-store", d)
	}
}

func TestEnsureFreshEmptyResources(t *testing.T) {
	d, err := EnsureFresh(context.Background(), nil, nil, Policy{StaleAfter: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	// nil db short-circuits first
	if d != DecisionNoStore {
		t.Fatalf("got %s want no-store", d)
	}
}
