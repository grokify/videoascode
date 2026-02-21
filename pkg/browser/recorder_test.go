package browser

import (
	"testing"
)

func TestBuildFrameDurations(t *testing.T) {
	// Create a recorder with mock data
	r := &Recorder{
		results: []StepResult{
			{
				Index: 0,
				Step: Step{
					Action:      ActionWait,
					MinDuration: 3000, // 3 seconds from TTS
				},
			},
			{
				Index: 1,
				Step: Step{
					Action:      ActionClick,
					MinDuration: 5000, // 5 seconds from TTS
				},
			},
			{
				Index: 2,
				Step: Step{
					Action:      ActionScroll,
					MinDuration: 0, // No voiceover for this step
				},
			},
			{
				Index: 3,
				Step: Step{
					Action:      ActionWait,
					MinDuration: 4000, // 4 seconds from TTS
				},
			},
		},
		timing: &TimingData{
			Steps: []StepTiming{
				{Index: 0, StartFrame: 0, EndFrame: 0},
				{Index: 1, StartFrame: 1, EndFrame: 1},
				{Index: 2, StartFrame: 2, EndFrame: 2},
				{Index: 3, StartFrame: 3, EndFrame: 3},
			},
		},
	}

	durations := r.buildFrameDurations()

	// Should have durations for frames 0, 1, 3 (steps with MinDuration > 0)
	if len(durations) != 3 {
		t.Errorf("expected 3 frame durations, got %d", len(durations))
	}

	// Verify each frame's duration
	expectedDurations := map[int]int{
		0: 3000, // Step 0's MinDuration
		1: 5000, // Step 1's MinDuration
		3: 4000, // Step 3's MinDuration
	}

	for frame, expected := range expectedDurations {
		if durations[frame] != expected {
			t.Errorf("frame %d: got duration %d, want %d", frame, durations[frame], expected)
		}
	}

	// Frame 2 should NOT have a duration (no MinDuration)
	if _, exists := durations[2]; exists {
		t.Errorf("frame 2 should not have a duration (step has no MinDuration)")
	}
}

func TestBuildFrameDurationsEmpty(t *testing.T) {
	r := &Recorder{
		results: []StepResult{},
		timing:  &TimingData{Steps: []StepTiming{}},
	}

	durations := r.buildFrameDurations()

	if len(durations) != 0 {
		t.Errorf("expected empty map for empty results, got %d entries", len(durations))
	}
}

func TestBuildFrameDurationsNoMinDuration(t *testing.T) {
	// All steps have MinDuration = 0 (no voiceovers)
	r := &Recorder{
		results: []StepResult{
			{Index: 0, Step: Step{Action: ActionClick, MinDuration: 0}},
			{Index: 1, Step: Step{Action: ActionScroll, MinDuration: 0}},
		},
		timing: &TimingData{
			Steps: []StepTiming{
				{Index: 0, StartFrame: 0, EndFrame: 0},
				{Index: 1, StartFrame: 1, EndFrame: 1},
			},
		},
	}

	durations := r.buildFrameDurations()

	if len(durations) != 0 {
		t.Errorf("expected empty map when no steps have MinDuration, got %d entries", len(durations))
	}
}

func TestBuildFrameDurationsMultipleFramesPerStep(t *testing.T) {
	// Step spans multiple frames
	r := &Recorder{
		results: []StepResult{
			{
				Index: 0,
				Step: Step{
					Action:      ActionWait,
					MinDuration: 6000, // 6 seconds
				},
			},
		},
		timing: &TimingData{
			Steps: []StepTiming{
				{Index: 0, StartFrame: 0, EndFrame: 2}, // 3 frames
			},
		},
	}

	durations := r.buildFrameDurations()

	// Duration should be split across frames: 6000 / 3 = 2000ms each
	expectedPerFrame := 2000
	for frame := 0; frame <= 2; frame++ {
		if durations[frame] != expectedPerFrame {
			t.Errorf("frame %d: got duration %d, want %d", frame, durations[frame], expectedPerFrame)
		}
	}
}
