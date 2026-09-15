package learncoroutine

import "testing"

// TestTasksRunOneStepPerFrame 验证等待一帧的任务从第 0 帧开始运行，
// 并且在每个逻辑帧中最多只执行一个步骤。任务第三次执行时主动结束，
// 测试同时确认结束后的任务会从待调度队列中移除。
func TestTasksRunOneStepPerFrame(t *testing.T) {
	s := New()
	// frames 依次记录任务每次实际运行时看到的逻辑帧，用于检查任务是否跨帧恢复。
	var frames []int64
	// steps 是任务内部保存的状态。闭包会在多次 Step 调用之间保留这个变量。
	steps := 0

	s.Spawn(func(ctx Context) StepResult {
		frames = append(frames, ctx.Frame)
		steps++
		if steps == 3 {
			// 第三步完成后退出任务，调度器不应再调用这个 Step。
			return Complete()
		}
		// 前两步各等待一帧，使任务在下一次 Tick 时恢复。
		return WaitFrames(1)
	})

	// 模拟宿主程序的帧循环，持续推进调度器，直到所有任务都执行完毕。
	for s.Pending() > 0 {
		s.Tick()
	}

	// 任务在第三次执行时完成，因此实际执行次数必须正好为 3。
	if len(frames) != 3 {
		t.Fatalf("steps = %d, want 3", len(frames))
	}
	for i, frame := range frames {
		// 每一步只等待一帧，所以三次执行的帧编号应依次为 0、1、2。
		if frame != int64(i) {
			t.Fatalf("step %d ran at frame %d, want %d", i, frame, i)
		}
	}
	// 返回 Complete 的任务必须被移除，不能继续占用待调度队列。
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
