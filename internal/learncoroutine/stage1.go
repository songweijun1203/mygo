// Package learncoroutine contains small, incremental coroutine exercises.
//
// Stage 1 is deliberately a state-machine scheduler. A task keeps its own
// state (usually in a closure or struct) and returns a result after each step.
package learncoroutine

// TaskID identifies a task. Stage 1 exposes IDs so later stages can add
// cancellation and joining without changing the creation API.
type TaskID uint64

// Context describes the scheduler state visible to one task step.
type Context struct {
	Frame int64
}

// StepResult tells the scheduler what to do after one task step.
type StepResult struct {
	// Done removes the task from the scheduler.
	Done bool
	// WaitFrames is the minimum number of future frames before the task runs
	// again. Non-positive values are normalized to one frame.
	WaitFrames int64
}

// Complete marks a task as finished.
func Complete() StepResult {
	return StepResult{Done: true}
}

// WaitFrames pauses a task for at least n future frames.
func WaitFrames(n int64) StepResult {
	if n < 1 {
		n = 1
	}
	return StepResult{WaitFrames: n}
}

// Step is one task's state-machine step. The function must do a bounded
// amount of work and return; a task that never returns blocks Tick.
type Step func(Context) StepResult

type task struct {
	id        TaskID
	step      Step
	wakeFrame int64
}

// Scheduler runs tasks cooperatively. It is intentionally single-threaded:
// the caller owns the scheduler and calls Tick from its event loop.
type Scheduler struct {
	frame  int64
	nextID TaskID

	tasks   []*task
	pending []*task
	ticking bool
}

// New creates an empty stage-one scheduler.
func New() *Scheduler {
	return &Scheduler{}
}

// Frame returns the current logical frame. A newly created scheduler starts at
// frame zero.
func (s *Scheduler) Frame() int64 {
	return s.frame
}

// Spawn adds a task. Tasks spawned while Tick is executing start on the next
// Tick, which makes one Tick a stable scheduling boundary.
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

// Pending reports how many tasks are waiting to run or resume.
func (s *Scheduler) Pending() int {
	return len(s.tasks) + len(s.pending)
}

// Tick executes at most one step for every task that is ready at the current
// frame, then advances the logical clock by one frame.
func (s *Scheduler) Tick() {
	if s.ticking {
		panic("learncoroutine: Tick called recursively")
	}

	s.ticking = true
	current := s.tasks
	s.tasks = nil

	for _, t := range current {
		if t.wakeFrame > s.frame {
			s.tasks = append(s.tasks, t)
			continue
		}

		result := t.step(Context{Frame: s.frame})
		if result.Done {
			continue
		}

		waitFrames := result.WaitFrames
		if waitFrames < 1 {
			waitFrames = 1
		}
		t.wakeFrame = s.frame + waitFrames
		s.tasks = append(s.tasks, t)
	}

	// New tasks are admitted only after the current frame has been drained.
	s.tasks = append(s.tasks, s.pending...)
	s.pending = nil
	s.ticking = false
	s.frame++
}
