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

// TestWaitFramesSkipsFutureFrames 验证 WaitFrames(2) 的等待语义。
// 任务在第 0 帧首次运行后应跳过第 1 帧，直到第 2 帧才获得下一次执行机会。
func TestWaitFramesSkipsFutureFrames(t *testing.T) {
	s := New()
	// 记录任务真正执行过的帧；任务处于等待状态的帧不会写入该切片。
	var frames []int64

	s.Spawn(func(ctx Context) StepResult {
		frames = append(frames, ctx.Frame)
		if len(frames) == 2 {
			// 第二次恢复后完成任务，测试只关注第一次等待产生的帧间隔。
			return Complete()
		}
		return WaitFrames(2)
	})

	// 新任务从第 0 帧开始执行。
	s.Tick()
	// 第一次 Tick 处理第 0 帧，任务应当立即执行。
	if got := frames[0]; got != 0 {
		t.Fatalf("first frame = %d, want 0", got)
	}
	// 第一次调用推进到第 1 帧，此时任务尚未到达 wakeFrame，不会执行。
	s.Tick()
	// 第二次调用处理第 2 帧，任务到达 wakeFrame 并执行第二步。
	s.Tick()
	if len(frames) != 2 || frames[1] != 2 {
		t.Fatalf("frames = %v, want [0 2]", frames)
	}
}

// TestSpawnDuringTickStartsNextTick 验证 Tick 执行期间创建的子任务不会立刻运行。
// 子任务应先进入 pending 队列，等当前帧的任务快照处理完毕后再转入 tasks，
// 因而最早只能在下一次 Tick 中执行。
func TestSpawnDuringTickStartsNextTick(t *testing.T) {
	s := New()
	// started 只由子任务修改，用来判断子任务在哪一次 Tick 中获得执行机会。
	started := false

	s.Spawn(func(Context) StepResult {
		// 父任务执行过程中创建子任务，此时调度器的 ticking 为 true。
		s.Spawn(func(Context) StepResult {
			started = true
			return Complete()
		})
		// 父任务创建完子任务后立即结束。
		return Complete()
	})

	// 第一次 Tick 只执行帧开始前已经存在的父任务，不能执行刚创建的子任务。
	s.Tick()
	if started {
		t.Fatal("child ran during the frame in which it was spawned")
	}
	// 当前帧结束后子任务已从 pending 转入 tasks，所以第二次 Tick 必须执行它。
	s.Tick()
	if !started {
		t.Fatal("child did not run on the next frame")
	}
}
