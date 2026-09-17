package main

import (
	"fmt"
	"time"

	"mygo/internal/learncoroutine"
)

func main() {
	scheduler := learncoroutine.New()

	scheduler.Spawn(func(ctx learncoroutine.Context) learncoroutine.StepResult {
		fmt.Println("A", ctx.Frame)
		return learncoroutine.WaitFrames(5)
	})

	scheduler.Spawn(func(ctx learncoroutine.Context) learncoroutine.StepResult {
		fmt.Println("B", ctx.Frame)
		return learncoroutine.Complete()
	})

	for scheduler.Pending() > 0 {
		scheduler.Tick()
		time.Sleep(33 * time.Millisecond)
	}
}
