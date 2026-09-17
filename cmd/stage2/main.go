package main

import (
	"fmt"
	"time"

	"mygo/internal/learncoroutine/stage2"
)

func main() {
	scheduler := stage2.New()

	scheduler.Spawn(func(ctx *stage2.Context) {
		for step := 1; step <= 3; step++ {
			fmt.Printf("frame=%d, step=%d\n", scheduler.Frame(), step)
			if step < 3 {
				ctx.Yield()
			}
		}
	})

	for scheduler.Pending() > 0 {
		scheduler.Tick()
		time.Sleep(33 * time.Millisecond)
	}
}
