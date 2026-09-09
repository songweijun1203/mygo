package main

import (
	"fmt"
	"time"

	"mygo/internal/learncoroutine"
)

func main() {
	scheduler := learncoroutine.New()
	count := 0

	scheduler.Spawn(func(ctx learncoroutine.Context) learncoroutine.StepResult {
		count++
		fmt.Printf("frame=%d, step=%d\n", ctx.Frame, count)
		if count == 3 {
			return learncoroutine.Complete()
		}
		return learncoroutine.WaitFrames(1)
	})

	for scheduler.Pending() > 0 {
		scheduler.Tick()
		time.Sleep(33 * time.Millisecond)
	}
}
