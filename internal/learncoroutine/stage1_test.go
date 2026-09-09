package learncoroutine

import "testing"

func TestTasksRunOneStepPerFrame(t *testing.T) {
	s := New()
	var frames []int64
	steps := 0

	s.Spawn(func(ctx Context) StepResult {
		frames = append(frames, ctx.Frame)
		steps++
		if steps == 3 {
			return Complete()
		}
		return WaitFrames(1)
	})

	for s.Pending() > 0 {
		s.Tick()
	}

	if len(frames) != 3 {
		t.Fatalf("steps = %d, want 3", len(frames))
	}
	for i, frame := range frames {
		if frame != int64(i) {
			t.Fatalf("step %d ran at frame %d, want %d", i, frame, i)
		}
	}
	if s.Pending() != 0 {
		t.Fatalf("pending = %d, want 0", s.Pending())
	}
}

func TestWaitFramesSkipsFutureFrames(t *testing.T) {
	s := New()
	var frames []int64

	s.Spawn(func(ctx Context) StepResult {
		frames = append(frames, ctx.Frame)
		if len(frames) == 2 {
			return Complete()
		}
		return WaitFrames(2)
	})

	s.Tick()
	if got := frames[0]; got != 0 {
		t.Fatalf("first frame = %d, want 0", got)
	}
	s.Tick()
	s.Tick()
	if len(frames) != 2 || frames[1] != 2 {
		t.Fatalf("frames = %v, want [0 2]", frames)
	}
}

func TestSpawnDuringTickStartsNextTick(t *testing.T) {
	s := New()
	started := false

	s.Spawn(func(Context) StepResult {
		s.Spawn(func(Context) StepResult {
			started = true
			return Complete()
		})
		return Complete()
	})

	s.Tick()
	if started {
		t.Fatal("child ran during the frame in which it was spawned")
	}
	s.Tick()
	if !started {
		t.Fatal("child did not run on the next frame")
	}
}
