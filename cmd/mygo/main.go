package main

import (
	"fmt"
	"time"

	coroutine "mygo/internal/corountine"
)

func work_thread(co *coroutine.Coroutines) {
	defer fmt.Println("work_thread 清理完成")
	var i int
	for {
		fmt.Printf("第 %d 帧: 协程正在运行\n", i)
		i++
		co.WaitNextFrame()
	}
}

func main() {
	coroutine.Start(nil)
	co := coroutine.New(func(report coroutine.PanicReport) {
		fmt.Printf("协程 %q 发生 panic: %v\n", report.Name, report.Value)
	})
	co.OnInited()

	//done := make(chan struct{})
	co.Create("counter", func(me coroutine.Thread) int {
		work_thread(co)
		return 0
	})

	fmt.Println("ready update")
	for i := 0; i < 10; i++ {
		co.Update()                               //推进协程调度器
		coroutine.Update(1, coroutine.DefaultFPS) //推进帧号和逻辑时间
		time.Sleep(1 * time.Second)
	}

	if co.AbortAllAndWait(time.Second) {
		fmt.Println("协程已完成")
	} else {
		fmt.Println("等待协程结束超时")
	}
}
