package toolevents

import (
	"errors"
	"testing"
	"time"
)

func TestDetectorPreservesStateOnInventoryErrorAndCompletesPassiveDeparture(t *testing.T) {
	tracker := NewTracker()
	sub := tracker.Subscribe()
	key := PaneKey{Session: "AGY", Window: 0, Pane: "%1"}
	detector := NewDetector(tracker, func() ([]PaneInfo, error) { return nil, errors.New("tmux unavailable") }, time.Second)
	detector.seen[key] = ToolAgy
	detector.detect()
	if got := detector.DetectedPanes()["%1"]; got != ToolAgy {
		t.Fatalf("state lost after inventory error: %q", got)
	}

	detector.listPane = func() ([]PaneInfo, error) { return []PaneInfo{}, nil }
	detector.detect()
	select {
	case event := <-sub:
		if event.Tool != ToolAgy || event.Status != StatusCompleted || !event.AutoDetected {
			t.Fatalf("unexpected departure event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected completed event")
	}
	if _, ok := detector.DetectedPanes()["%1"]; ok {
		t.Fatal("departure must clear detector ownership")
	}
	if len(tracker.GetAll()) != 0 {
		t.Fatal("completed departure must clear state projection")
	}
}
