// Package learncoroutine 提供一组循序渐进的协作式协程练习。
//
// 第一阶段使用状态机实现一个单线程调度器。每个任务需要自行保存执行状态，
// 状态通常保存在闭包捕获的变量或结构体中。调度器每次调用任务时只执行一个步骤，
// 任务返回 StepResult，明确告诉调度器自己已经完成，或者需要等待多少逻辑帧后继续。
//
// 这里的任务不同于 Go 原生 goroutine：它不会被 Go 运行时抢占调度，也不会并行执行。
// 任务只有在 Step 返回后才会让出执行权，因此这是一种由任务主动配合的协作式调度。
package learncoroutine

// TaskID 是调度器为任务分配的唯一标识。
// 第一阶段暂时只负责返回 ID；提前公开该类型，是为了后续加入取消任务、等待任务结束等能力时，
// 不必修改 Spawn 的调用方式。
type TaskID uint64

// Context 保存任务执行当前步骤时能够观察到的调度器状态。
// Context 按值传入 Step，任务只能读取这次执行对应的上下文，不能借此直接修改调度器。
type Context struct {
	// Frame 是当前逻辑帧编号。新建调度器从第 0 帧开始，每次 Tick 结束后递增 1。
	Frame int64
}

// StepResult 描述一个任务步骤执行完毕后，调度器应当如何处理该任务。
type StepResult struct {
	// Done 表示任务是否已经全部完成。值为 true 时，调度器会移除该任务，且不会再次调用它。
	Done bool
	// WaitFrames 表示任务至少需要等待多少个逻辑帧才能再次执行。
	// 当 Done 为 false 时，非正数会被统一修正为 1，避免同一个任务在一次 Tick 中重复执行。
	WaitFrames int64
}

// Complete 返回一个表示任务已经完成的结果。
// 任务返回该结果后会立即从调度器中移除，不再参与后续 Tick。
func Complete() StepResult {
	return StepResult{Done: true}
}

// WaitFrames 返回一个等待结果，使任务至少经过 n 个逻辑帧后再恢复执行。
// n 小于 1 时按 1 处理，因此调用 WaitFrames(0) 也会让任务至少让出当前帧。
func WaitFrames(n int64) StepResult {
	if n < 1 {
		n = 1
	}
	return StepResult{WaitFrames: n}
}

// Step 表示任务状态机的一次执行步骤。
//
// Step 应当只完成有限量的工作并及时返回。调度器是单线程且协作式的，
// 如果某个 Step 长时间阻塞或者永不返回，当前 Tick 和其他所有任务都会被一起阻塞。
type Step func(Context) StepResult

// task 保存调度器内部使用的任务状态，对包外不可见。
type task struct {
	// id 是创建任务时分配的唯一标识。
	id        TaskID
	// step 是任务下一次获得调度机会时需要执行的函数。
	step      Step
	// wakeFrame 是允许任务再次执行的最早逻辑帧。
	wakeFrame int64
}

// Scheduler 以协作方式调度所有任务。
//
// Scheduler 特意设计为单线程模型：调用方持有调度器，并在自己的事件循环或帧循环中调用 Tick。
// Tick 会按任务进入队列的顺序依次执行就绪任务，每个任务在一帧内至多执行一个步骤。
// Scheduler 本身没有提供并发访问保护，不应由多个 goroutine 同时调用其方法。
type Scheduler struct {
	// frame 是当前逻辑帧编号。
	frame  int64
	// nextID 保存最近一次分配的任务 ID，新任务使用递增后的值。
	nextID TaskID

	// tasks 保存已经进入调度器、等待执行或恢复的任务。
	tasks   []*task
	// pending 暂存 Tick 执行期间创建的新任务，防止它们在当前帧立即运行。
	pending []*task
	// ticking 用于标记 Tick 正在执行，同时阻止递归调用 Tick。
	ticking bool
}

// New 创建一个不包含任何任务的第一阶段调度器。
// 新调度器的逻辑帧为 0，首次创建的任务 ID 为 1。
func New() *Scheduler {
	return &Scheduler{}
}

// Frame 返回调度器当前的逻辑帧编号。
// 新建调度器从第 0 帧开始；每次 Tick 完成所有本帧工作后，帧编号增加 1。
func (s *Scheduler) Frame() int64 {
	return s.frame
}

// Spawn 向调度器添加一个任务，并返回该任务的唯一 ID。
//
// 在 Tick 之外创建的任务会立即进入任务队列，可以在下一次 Tick 中执行。
// 在 Tick 执行期间创建的任务会先进入 pending 队列，直到当前帧的已有任务全部处理完毕后才转入
// tasks 队列，因此最早也只能在下一次 Tick 中执行。这个规则保证了一次 Tick 的任务集合是稳定的。
// step 为 nil 表示调用方无法提供有效任务，Spawn 会直接触发 panic。
func (s *Scheduler) Spawn(step Step) TaskID {
	if step == nil {
		panic("learncoroutine: nil task step")
	}

	s.nextID++
	t := &task{id: s.nextID, step: step, wakeFrame: s.frame}
	if s.ticking {
		s.pending = append(s.pending, t)
	} else {
		s.tasks = append(s.tasks, t)
	}
	return t.id
}

// Pending 返回尚未完成的任务数量。
// 结果同时包含 tasks 中等待执行或恢复的任务，以及 Tick 期间新建、仍在 pending 中的任务。
func (s *Scheduler) Pending() int {
	return len(s.tasks) + len(s.pending)
}

// Tick 推进调度器一个逻辑帧。
//
// Tick 会取得当前任务队列的稳定快照，并按照入队顺序检查每个任务。已经到达唤醒帧的任务
// 最多执行一个 Step；尚未到达唤醒帧的任务会保留在队列中。步骤执行完成的任务会被移除，
// 仍需继续的任务则根据 WaitFrames 计算下一次可运行的帧。全部任务处理完毕后，Tick 接纳
// 当前帧中新建的任务，最后把逻辑帧编号增加 1。
//
// Tick 不允许递归调用，因为递归推进帧会破坏当前任务快照和 pending 队列的边界。
func (s *Scheduler) Tick() {
	if s.ticking {
		panic("learncoroutine: Tick called recursively")
	}

	s.ticking = true
	// 先取出本帧开始时已经存在的任务，确保遍历期间新建的任务不会混入当前快照。
	current := s.tasks
	s.tasks = nil

	for _, t := range current {
		// 未到唤醒帧的任务保持等待状态，本帧不调用它的 Step。
		if t.wakeFrame > s.frame {
			s.tasks = append(s.tasks, t)
			continue
		}

		result := t.step(Context{Frame: s.frame})
		// 已完成的任务不再放回任务队列，相当于从调度器中移除。
		if result.Done {
			continue
		}

		// 所有未完成任务至少让出一个逻辑帧，保证每个任务一帧内至多执行一步。
		waitFrames := result.WaitFrames
		if waitFrames < 1 {
			waitFrames = 1
		}
		t.wakeFrame = s.frame + waitFrames
		s.tasks = append(s.tasks, t)
	}

	// 当前帧的任务全部处理完毕后，再接纳本帧执行期间创建的新任务。
	s.tasks = append(s.tasks, s.pending...)
	s.pending = nil
	s.ticking = false
	s.frame++
}
