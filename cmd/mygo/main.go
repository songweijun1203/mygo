package main

import (
	"fmt"
	"time"

	coroutine "mygo/internal/corountine"
)

func main() {
	coroutine.Start(nil)
	co := coroutine.New(func(report coroutine.PanicReport) {
		fmt.Printf("协程 %q 发生 panic: %v\n", report.Name, report.Value)
	})
	co.OnInited()

	done := make(chan struct{})
	co.Create("counter", func(me coroutine.Thread) int {
		//for i := 1; i <= 3; i++ {
		fmt.Printf("第 %d 帧: 协程正在运行\n", 1)
		co.WaitNextFrame()
		fmt.Printf("第 %d 帧: 协程正在运行\n", 2)
		co.WaitNextFrame()
		fmt.Printf("第 %d 帧: 协程正在运行\n", 3)
		co.WaitNextFrame()
		//}
		close(done)
		return 0
	})

	for {
		fmt.Println("ready update")
		co.Update()
		select {
		case <-done:
			co.AbortAllAndWait(time.Second)
			fmt.Println("协程已完成")
			return
		default:
		}
		coroutine.Update(1.0/30.0, coroutine.DefaultFPS)
		time.Sleep(10 * time.Millisecond)
	}
}
