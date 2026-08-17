package overlay

import (
	"encoding/json"
	"testing"
)

func TestOverlayPlanAndFchg(t *testing.T) {
	plan := json.RawMessage(`{"timetable":{"@station":"Frankfurt(Main)Hbf","s":[
		{"@id":"1","tl":{"@c":"ICE","@n":"228"},"dp":{"@pt":"2608171341","@pp":"6","@ppth":"Mainz Hbf|Köln Hbf"}},
		{"@id":"2","tl":{"@c":"RE","@n":"50"},"dp":{"@pt":"2608171350","@pp":"12"}}
	]}}`)
	fchg := json.RawMessage(`{"timetable":{"@station":"Frankfurt(Main)Hbf","s":[
		{"@id":"1","dp":{"@ct":"2608171348","@cp":"7"}},
		{"@id":"2","dp":{"@cs":"c","@clt":"2608171320"}}
	]}}`)
	_, planned, err := ParseTimetable(plan)
	if err != nil {
		t.Fatal(err)
	}
	_, changes, err := ParseTimetable(fchg)
	if err != nil {
		t.Fatal(err)
	}
	stops := Overlay(planned, changes)
	if len(stops) != 2 {
		t.Fatalf("stops=%d", len(stops))
	}
	byID := map[string]Stop{}
	for _, s := range stops {
		byID[s.ID] = s
	}
	ice := byID["1"]
	if ice.Train != "ICE 228" || ice.Platform != "7" || !ice.PlatformChanged {
		t.Fatalf("ice overlay: %+v", ice)
	}
	if ice.DelayMinutes == nil || *ice.DelayMinutes != 7 {
		t.Fatalf("ice delay=%v", ice.DelayMinutes)
	}
	if ice.Destination != "Köln Hbf" {
		t.Fatalf("dest=%q", ice.Destination)
	}
	re := byID["2"]
	if !re.Cancelled {
		t.Fatalf("expected cancel: %+v", re)
	}
	plats := PlatformChanges(stops)
	if len(plats) != 1 || plats[0].ID != "1" {
		t.Fatalf("platforms=%+v", plats)
	}
	cancels := Cancellations(stops)
	if len(cancels) != 1 || cancels[0].ID != "2" {
		t.Fatalf("cancels=%+v", cancels)
	}
	delays := Delays(stops)
	if len(delays) != 1 || delays[0].ID != "1" {
		t.Fatalf("delays=%+v", delays)
	}
}

func TestRecentOnly(t *testing.T) {
	stops := []Stop{{ID: "1", Train: "ICE 1"}, {ID: "2", Train: "RE 2"}}
	recent := map[string]Stop{"2": {ID: "2"}}
	marked := MarkRecent(stops, recent)
	only := RecentOnly(marked)
	if len(only) != 1 || only[0].ID != "2" || !only[0].Recent {
		t.Fatalf("recent=%+v", only)
	}
}
