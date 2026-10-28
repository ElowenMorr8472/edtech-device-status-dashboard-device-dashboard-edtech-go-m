package main

import (
	"testing"
	"time"
)

func TestDecideStatusKeepsLearnerOnTrackBeforeDeadline(t *testing.T) {
	now := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	input := deviceStatus{Online: true, ProgressPct: 70, Deadline: now.Add(time.Hour)}
	if got := decideStatus(input, now); got != "on_track" {
		t.Fatalf("decision = %q, want on_track", got)
	}
	input.Deadline = now.Add(-time.Minute)
	if got := decideStatus(input, now); got != "needs_attention" {
		t.Fatalf("expired deadline decision = %q, want needs_attention", got)
	}
}
