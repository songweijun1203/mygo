# 第一阶段：手写状态机调度器

`internal/learncoroutine` 是学习用实现，和现有的 `internal/corountine` 相互独立。

这一阶段不使用 Go goroutine，也不保存函数调用栈。每个任务通过闭包或结构体保存自己的状态，每次 `Scheduler.Tick` 最多执行一步，然后返回一个 `StepResult`：

- `Complete()`：任务结束。
- `WaitFrames(n)`：至少等待 n 个逻辑帧后再次执行。

运行示例：

```bash
go run ./cmd/stage1
```

这个版本要先验证的是调度语义：任务可以被注册、按帧轮流执行、延迟恢复，并且在完成后从调度器移除。它还不是栈式协程，因此用户必须自己把状态保存在闭包或结构体中。
