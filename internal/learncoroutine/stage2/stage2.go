// Package stage2 演示如何使用 Go goroutine 和通道实现可暂停、可恢复的顺序代码。
package stage2

import "fmt"

// TaskID 是调度器为任务分配的唯一标识。
type TaskID uint64

type event uint8

const (
	eventYielded event = iota
	eventCompleted
)

// Context 是任务与调度器协作的入口。
// 每个 Context 只属于创建它的任务，不应传递给其他 goroutine 使用。
type Context struct {
	resume <-chan struct{}
	events chan<- event
}

// Yield 主动让出执行权，并阻塞当前任务，直到后续 Tick 再次恢复它。
// 调用 Yield 时，当前函数的局部变量和完整调用栈都会保留。
func (c *Context) Yield() {
	//这个函数在具体的协程函数内部运行，所以阻塞的是当前协程
	c.events <- eventYielded //通知调度器。我已经挂起
	<-c.resume               //等待被恢复
}

// Task 是可以主动暂停和恢复的任务函数。
type Task func(*Context)

type coroutine struct {
	id     TaskID
	resume chan struct{}
	events chan event
}

func newCoroutine(id TaskID, fn Task) *coroutine {
	c := &coroutine{
		id:     id,
		resume: make(chan struct{}),
		events: make(chan event),
	}
	//创建一个上下文，将协程的必要信息 放入上下文中，这样协程函数运行时候
	//可以拿到上下文信息
	ctx := &Context{resume: c.resume, events: c.events}

	go func() {
		// 新任务必须先等到 Tick 发出第一次恢复信号，不能在 Spawn 中直接运行。
		<-c.resume
		fn(ctx)
		//通知调度器
		c.events <- eventCompleted
	}()

	return c
}

// Scheduler 以协作方式调度多个任务。
// Scheduler 不支持并发调用；运行中的任务可以调用 Spawn，但外部 goroutine
// 不应在 Tick 执行期间操作同一个 Scheduler。
type Scheduler struct {
	frame  int64
	nextID TaskID

	tasks   []*coroutine
	pending []*coroutine
	ticking bool
}

// New 创建一个空的第二阶段调度器。
func New() *Scheduler {
	return &Scheduler{}
}

// Frame 返回当前逻辑帧编号。每次 Tick 完成后，帧编号增加 1。
func (s *Scheduler) Frame() int64 {
	return s.frame
}

// Spawn 创建一个尚未开始执行的任务，并返回任务 ID。
// 在 Tick 期间创建的任务最早会在下一次 Tick 中运行。
func (s *Scheduler) Spawn(fn Task) TaskID {
	if fn == nil {
		panic("stage2: nil task")
	}

	s.nextID++
	c := newCoroutine(s.nextID, fn)
	if s.ticking {
		s.pending = append(s.pending, c)
	} else {
		s.tasks = append(s.tasks, c)
	}
	return c.id
}

// Pending 返回尚未完成的任务数量。
func (s *Scheduler) Pending() int {
	return len(s.tasks) + len(s.pending)
}

// Tick 让当前已有的每个任务依次运行，直到该任务 Yield 或执行结束。
// 一个任务如果既不调用 Yield 也不返回，就会阻塞本次 Tick 和其他所有任务。
func (s *Scheduler) Tick() {
	if s.ticking {
		panic("stage2: Tick called recursively")
	}

	s.ticking = true
	current := s.tasks
	s.tasks = nil

	//循环唤醒每一个协程
	for _, c := range current {
		c.resume <- struct{}{} //唤醒协程
		//阻塞等待协程挂起或者完成

		switch <-c.events {
		case eventYielded:
			fmt.Println("调度器已经检测到协程挂起")
			s.tasks = append(s.tasks, c)
		case eventCompleted:
			fmt.Println("调度器已经检测到协程结束")
		}
	}

	//将tick期间加进来的任务 移动到当前任务队列，下一次推进逻辑帧时候调度
	s.tasks = append(s.tasks, s.pending...)
	s.pending = nil
	s.ticking = false
	s.frame++ //逻辑帧递增
}
