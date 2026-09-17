# 第二阶段：可以暂停的顺序代码

第二阶段位于 `internal/learncoroutine/stage2`，它在第一阶段的状态机调度概念之上，引入 Go goroutine 和无缓冲通道。

## 本阶段目标

第一阶段的任务每执行一步都必须返回，状态由调用者手动保存在闭包或结构体中。第二阶段让任务写成普通顺序代码，并通过 `ctx.Yield()` 在函数内部暂停：

```go
scheduler.Spawn(func(ctx *stage2.Context) {
	local := 42
	fmt.Println("第一步")
	ctx.Yield()
	fmt.Println("第二步", local)
})
```

恢复后会从 `Yield` 的下一行继续，局部变量和嵌套函数调用栈仍然存在。这不是 Go 编译器生成的协程状态机，而是利用一个被阻塞的 Go goroutine 保存真实调用栈。

## 运行方式

```bash
go run ./cmd/stage2
go test ./internal/learncoroutine/stage2
```

## 调度握手

每个任务拥有两个无缓冲通道：

- `resume`：调度器通知任务开始或继续执行。
- `events`：任务通知调度器自己已经 `Yield` 或执行完毕。

一次调度过程如下：

```text
Scheduler.Tick
    -> 向 resume 发送信号
    -> 任务继续执行
    -> 任务调用 Yield
    -> 向 events 发送 yielded
    -> 任务等待下一次 resume
    -> Tick 处理下一个任务
```

如果任务函数直接返回，它会发送 completed，调度器随后将其移除。

## 当前限制

- 只有主动 `Yield`，还不能指定等待几帧或等待时间。
- 没有取消、Join、channel 等待和 panic 隔离。
- Scheduler 不是并发安全的。
- 放弃一个仍在暂停的 Scheduler 会留下等待恢复的 goroutine。
- 任务在两次 `Yield` 之间必须做有限工作，否则整个 Tick 会一直等待它。

下一阶段可以给 `Yield` 事件增加原因和唤醒条件，从而重新加入 `WaitFrames(n)`，但保持顺序式任务代码不变。
