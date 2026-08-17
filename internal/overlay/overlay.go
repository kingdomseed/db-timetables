// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// Overlay plan + fchg/rchg IRIS XML-as-JSON into a traveler-shaped station board.

package overlay

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Stop is one planned stop at a station after live changes are applied.
type Stop struct {
	ID                string `json:"id"`
	Train             string `json:"train"`
	Category          string `json:"category,omitempty"`
	Number            string `json:"number,omitempty"`
	Station           string `json:"station,omitempty"`
	PlannedTime       string `json:"planned_time,omitempty"`
	ChangedTime       string `json:"changed_time,omitempty"`
	DelayMinutes      *int   `json:"delay_minutes,omitempty"`
	PlannedPlatform   string `json:"planned_platform,omitempty"`
	ChangedPlatform   string `json:"changed_platform,omitempty"`
	Platform          string `json:"platform,omitempty"`
	PlatformChanged   bool   `json:"platform_changed"`
	Cancelled         bool   `json:"cancelled"`
	Recent            bool   `json:"recent,omitempty"`
	Destination       string `json:"destination,omitempty"`
	ArrivalPlanned    string `json:"arrival_planned,omitempty"`
	DeparturePlanned  string `json:"departure_planned,omitempty"`
	plannedIRIS       string
	changedIRIS       string
}

// Board is the overlaid station picture for one hour.
type Board struct {
	EVA     string `json:"eva"`
	Date    string `json:"date"`
	Hour    string `json:"hour"`
	Station string `json:"station,omitempty"`
	Stops   []Stop `json:"stops"`
}

type event struct {
	plannedTime     string
	changedTime     string
	plannedPlatform string
	changedPlatform string
	status          string
	cancelTime      string
	path            string
}

// ParseTimetable turns XMLToJSON timetable JSON into stops keyed by stop id.
func ParseTimetable(raw json.RawMessage) (station string, stops map[string]Stop, err error) {
	stops = map[string]Stop{}
	if len(raw) == 0 {
		return "", stops, nil
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", nil, fmt.Errorf("parse timetable json: %w", err)
	}
	tt := asMap(root["timetable"])
	if tt == nil {
		// sometimes the document is the timetable object itself
		if _, ok := root["s"]; ok {
			tt = root
		}
	}
	if tt == nil {
		return "", stops, nil
	}
	station = attr(tt, "station")
	for _, item := range asList(tt["s"]) {
		m := asMap(item)
		if m == nil {
			continue
		}
		stop := stopFromMap(m, station)
		if stop.ID == "" {
			continue
		}
		stops[stop.ID] = stop
	}
	return station, stops, nil
}

// Overlay applies change stops onto planned stops. Changes win for time/platform/cancel.
func Overlay(plan, changes map[string]Stop) []Stop {
	out := make(map[string]Stop, len(plan)+len(changes))
	for id, s := range plan {
		out[id] = s
	}
	for id, ch := range changes {
		base, ok := out[id]
		if !ok {
			out[id] = ch
			continue
		}
		out[id] = mergeStop(base, ch)
	}
	rows := make([]Stop, 0, len(out))
	for _, s := range out {
		rows = append(rows, s)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].sortKey() < rows[j].sortKey()
	})
	return rows
}

// MarkRecent flags stops whose ids appear in the rchg set.
func MarkRecent(stops []Stop, recent map[string]Stop) []Stop {
	if len(recent) == 0 {
		return stops
	}
	out := make([]Stop, 0, len(stops)+len(recent))
	seen := map[string]bool{}
	for _, s := range stops {
		if _, ok := recent[s.ID]; ok {
			s.Recent = true
		}
		out = append(out, s)
		seen[s.ID] = true
	}
	for id, s := range recent {
		if seen[id] {
			continue
		}
		s.Recent = true
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].sortKey() < out[j].sortKey()
	})
	return out
}

func PlatformChanges(stops []Stop) []Stop {
	rows := make([]Stop, 0)
	for _, s := range stops {
		if s.PlatformChanged {
			rows = append(rows, s)
		}
	}
	return rows
}

func Cancellations(stops []Stop) []Stop {
	rows := make([]Stop, 0)
	for _, s := range stops {
		if s.Cancelled {
			rows = append(rows, s)
		}
	}
	return rows
}

func Delays(stops []Stop) []Stop {
	rows := make([]Stop, 0)
	for _, s := range stops {
		if s.Cancelled {
			continue
		}
		if s.DelayMinutes != nil && *s.DelayMinutes > 0 {
			rows = append(rows, s)
		}
	}
	return rows
}

func RecentOnly(stops []Stop) []Stop {
	rows := make([]Stop, 0)
	for _, s := range stops {
		if s.Recent {
			rows = append(rows, s)
		}
	}
	return rows
}

func (s Stop) sortKey() string {
	if s.PlannedTime != "" {
		return s.PlannedTime + s.ID
	}
	if s.ChangedTime != "" {
		return s.ChangedTime + s.ID
	}
	return s.ID
}

func stopFromMap(m map[string]any, station string) Stop {
	id := attr(m, "id")
	tl := asMap(first(m["tl"]))
	cat := attr(tl, "c")
	num := attr(tl, "n")
	train := strings.TrimSpace(cat + " " + num)
	ar := parseEvent(asMap(first(m["ar"])))
	dp := parseEvent(asMap(first(m["dp"])))
	primary := dp
	if primary.plannedTime == "" && primary.changedTime == "" {
		primary = ar
	}
	plannedIRIS := firstNonEmpty(primary.plannedTime, ar.plannedTime, dp.plannedTime)
	changedIRIS := firstNonEmpty(primary.changedTime, ar.changedTime, dp.changedTime)
	s := Stop{
		ID:               id,
		Train:            train,
		Category:         cat,
		Number:           num,
		Station:          station,
		PlannedTime:      formatClock(plannedIRIS),
		ChangedTime:      formatClock(changedIRIS),
		PlannedPlatform:  firstNonEmpty(primary.plannedPlatform, ar.plannedPlatform, dp.plannedPlatform),
		ChangedPlatform:  firstNonEmpty(primary.changedPlatform, ar.changedPlatform, dp.changedPlatform),
		Cancelled:        isCancelled(ar) || isCancelled(dp),
		Destination:      lastPath(firstNonEmpty(dp.path, ar.path)),
		ArrivalPlanned:   formatClock(ar.plannedTime),
		DeparturePlanned: formatClock(dp.plannedTime),
	}
	s.Platform = firstNonEmpty(s.ChangedPlatform, s.PlannedPlatform)
	s.PlatformChanged = s.ChangedPlatform != "" && s.ChangedPlatform != s.PlannedPlatform
	if d, ok := delayMinutes(plannedIRIS, changedIRIS); ok {
		s.DelayMinutes = &d
	}
	s.plannedIRIS = plannedIRIS
	s.changedIRIS = changedIRIS
	return s
}

func mergeStop(base, ch Stop) Stop {
	out := base
	if ch.Train != "" {
		out.Train = ch.Train
		out.Category = ch.Category
		out.Number = ch.Number
	}
	if ch.ChangedTime != "" {
		out.ChangedTime = ch.ChangedTime
	}
	if ch.ChangedPlatform != "" {
		out.ChangedPlatform = ch.ChangedPlatform
	}
	if ch.PlannedTime != "" && out.PlannedTime == "" {
		out.PlannedTime = ch.PlannedTime
	}
	if ch.PlannedPlatform != "" && out.PlannedPlatform == "" {
		out.PlannedPlatform = ch.PlannedPlatform
	}
	if ch.Destination != "" {
		out.Destination = ch.Destination
	}
	out.Cancelled = out.Cancelled || ch.Cancelled
	out.Platform = firstNonEmpty(out.ChangedPlatform, out.PlannedPlatform)
	out.PlatformChanged = out.ChangedPlatform != "" && out.ChangedPlatform != out.PlannedPlatform
	if ch.plannedIRIS != "" && out.plannedIRIS == "" {
		out.plannedIRIS = ch.plannedIRIS
	}
	if ch.changedIRIS != "" {
		out.changedIRIS = ch.changedIRIS
	}
	if d, ok := delayMinutes(out.plannedIRIS, out.changedIRIS); ok {
		out.DelayMinutes = &d
	} else if ch.DelayMinutes != nil {
		out.DelayMinutes = ch.DelayMinutes
	}
	out.Recent = out.Recent || ch.Recent
	return out
}

func parseEvent(m map[string]any) event {
	if m == nil {
		return event{}
	}
	return event{
		plannedTime:     attr(m, "pt"),
		changedTime:     attr(m, "ct"),
		plannedPlatform: attr(m, "pp"),
		changedPlatform: attr(m, "cp"),
		status:          attr(m, "cs"),
		cancelTime:      attr(m, "clt"),
		path:            firstNonEmpty(attr(m, "cpth"), attr(m, "ppth")),
	}
}

func isCancelled(e event) bool {
	return e.status == "c" || e.cancelTime != ""
}

func delayMinutes(planned, changed string) (int, bool) {
	pt, ok1 := parseIRISTime(planned)
	ct, ok2 := parseIRISTime(changed)
	if !ok1 || !ok2 {
		return 0, false
	}
	return int(ct.Sub(pt).Minutes()), true
}

func parseIRISTime(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if len(v) != 10 {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("0601021504", v, berlin())
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func formatClock(v string) string {
	t, ok := parseIRISTime(v)
	if !ok {
		return v
	}
	return t.Format("15:04")
}

func parseClockBack(v string) string {
	// already IRIS
	if len(v) == 10 {
		return v
	}
	return ""
}

func berlin() *time.Location {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return time.FixedZone("CET", 2*3600)
	}
	return loc
}

// NowSlice returns today's YYMMDD and current hour HH in Europe/Berlin.
func NowSlice(now time.Time) (date, hour string) {
	t := now.In(berlin())
	return t.Format("060102"), t.Format("15")
}

func lastPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "|")
	return strings.TrimSpace(parts[len(parts)-1])
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func attr(m map[string]any, name string) string {
	if m == nil {
		return ""
	}
	if v, ok := m["@"+name]; ok {
		return stringify(v)
	}
	if v, ok := m[name]; ok {
		return stringify(v)
	}
	return ""
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asList(v any) []any {
	if v == nil {
		return nil
	}
	if list, ok := v.([]any); ok {
		return list
	}
	return []any{v}
}

func first(v any) any {
	list := asList(v)
	if len(list) == 0 {
		return nil
	}
	return list[0]
}
